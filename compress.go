package hblade

import (
	"io"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
)

// Response content encodings supported by SetCompression.
const (
	EncodingZstd = "zstd"
	EncodingGzip = "gzip"

	// RFC 9659: HTTP clients (browsers) only have to support zstd windows up to 8MB.
	zstdMaxWindow = 8 << 20

	// Bigger zstd output buffers are dropped instead of pooled to keep idle memory low.
	maxPooledZstdBuf = 1 << 20
)

var contentEncodingZstdValue = []string{EncodingZstd}

type compression struct {
	encodings []string // server preference order; empty (default) disables compression

	gzipLevel int
	gzipPool  sync.Pool

	zstdLevel zstd.EncoderLevel
	zstdOnce  sync.Once
	zstdEnc   *zstd.Encoder
	zstdBufs  sync.Pool
}

func (c *compression) init() {
	c.gzipLevel = gzip.DefaultCompression
	c.gzipPool.New = c.newGzipWriter
	c.zstdLevel = zstd.SpeedFastest
}

func (c *compression) newGzipWriter() any {
	w, _ := gzip.NewWriterLevel(io.Discard, c.gzipLevel)
	return w
}

// The encoder is created lazily so apps that never compress don't pay for its tables.
func (c *compression) zstdEncoder() *zstd.Encoder {
	c.zstdOnce.Do(func() {
		c.zstdEnc, _ = zstd.NewWriter(nil,
			zstd.WithEncoderLevel(c.zstdLevel),
			zstd.WithWindowSize(zstdMaxWindow),
			zstd.WithLowerEncoderMem(true),
		)
	})
	return c.zstdEnc
}

// negotiate returns the preferred encoding accepted by the client, or "".
func (c *compression) negotiate(acceptEncoding string) string {
	if acceptEncoding == "" {
		return ""
	}
	for _, enc := range c.encodings {
		if acceptsEncoding(acceptEncoding, enc) {
			return enc
		}
	}
	return ""
}

func (c *compression) writeGzip(w io.Writer, body []byte) error {
	gz := c.gzipPool.Get().(*gzip.Writer)
	gz.Reset(w)
	_, err := gz.Write(body)
	if cerr := gz.Close(); err == nil {
		err = cerr
	}
	gz.Reset(io.Discard)
	c.gzipPool.Put(gz)
	return err
}

// encodeZstd compresses body into a pooled buffer; release it with releaseZstd.
func (c *compression) encodeZstd(body []byte) *[]byte {
	bp, _ := c.zstdBufs.Get().(*[]byte)
	if bp == nil {
		bp = new([]byte)
	}
	*bp = c.zstdEncoder().EncodeAll(body, (*bp)[:0])
	return bp
}

func (c *compression) releaseZstd(bp *[]byte) {
	if cap(*bp) <= maxPooledZstdBuf {
		c.zstdBufs.Put(bp)
	}
}

// acceptsEncoding reports whether an Accept-Encoding header value allows name,
// honoring explicit rejections such as "gzip;q=0".
func acceptsEncoding(header, name string) bool {
	for header != "" {
		var part string
		part, header, _ = strings.Cut(header, ",")
		token, params, _ := strings.Cut(part, ";")
		if strings.EqualFold(strings.TrimSpace(token), name) {
			return !qIsZero(params)
		}
	}
	return false
}

func qIsZero(params string) bool {
	for params != "" {
		var p string
		p, params, _ = strings.Cut(params, ";")
		k, v, ok := strings.Cut(p, "=")
		if ok && strings.EqualFold(strings.TrimSpace(k), "q") {
			q, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
			return err == nil && q == 0
		}
	}
	return false
}

// SetCompression enables response compression with the given encodings in server
// preference order, e.g. SetCompression(EncodingZstd, EncodingGzip). Unknown values
// are ignored; call it with no arguments to disable compression again.
// Compression is disabled by default since nginx/CDN usually handles it.
// Call it before serving requests.
func (b *Blade) SetCompression(encodings ...string) {
	list := make([]string, 0, len(encodings))
	for _, e := range encodings {
		if (e == EncodingZstd || e == EncodingGzip) && !slices.Contains(list, e) {
			list = append(list, e)
		}
	}
	b.compression.encodings = list
}

// SetGzipLevel sets the gzip compression level (gzip.BestSpeed..gzip.BestCompression,
// gzip.DefaultCompression or gzip.HuffmanOnly). Default is gzip.DefaultCompression.
// Call it before serving requests.
func (b *Blade) SetGzipLevel(level int) {
	if level < gzip.HuffmanOnly || level > gzip.BestCompression {
		return
	}
	c := &b.compression
	c.gzipLevel = level
	c.gzipPool = sync.Pool{New: c.newGzipWriter}
}

// SetZstdLevel sets the zstd level using standard zstd levels (1-22), mapped to
// the closest supported encoder speed. Default is the fastest level (~zstd 1),
// which still compresses about as well as gzip level 6 at a fraction of the CPU.
// Call it before serving requests.
func (b *Blade) SetZstdLevel(level int) {
	c := &b.compression
	c.zstdLevel = zstd.EncoderLevelFromZstd(level)
	c.zstdOnce = sync.Once{}
	c.zstdEnc = nil
}
