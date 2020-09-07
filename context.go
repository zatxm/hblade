package hblade

import (
	stdContext "context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/zatxm/hblade/binding"
	"github.com/zatxm/hblade/internal/bytesconv"
)

const (
	// 此值应接近TCP包大小,设置为256将最终数据包的大小减少了约70个字节
	gzipThreshold = 256

	// 路由最大参数
	maxParams = 64
)

// Context represents a request & response context.
type Context struct {
	b             *Blade
	status        int
	request       request
	response      response
	handler       Handler
	paramNames    [maxParams]string
	paramValues   [maxParams]string
	paramCount    int
	modifierCount int
	sameSite      http.SameSite
}

// 返回blade
func (c *Context) B() *Blade {
	return c.b
}

// 返回字节处理
func (c *Context) Bytes(body []byte) error {
	// If the request has been canceled by the client, stop.
	if c.request.Context().Err() != nil {
		return errors.New("Request interrupted by the client")
	}

	// Small response
	if len(body) < gzipThreshold {
		c.response.inner.WriteHeader(c.status)
		_, err := c.response.inner.Write(body)
		return err
	}

	// Content type
	header := c.response.inner.Header()
	contentType := header.Get(contentTypeHeader)
	isMediaType := isMedia(contentType)

	// Cache control header
	if isMediaType {
		header.Set(cacheControlHeader, cacheControlMedia)
	} else {
		header.Set(cacheControlHeader, cacheControlAlwaysValidate)
	}

	// No GZip?
	clientSupportsGZip := strings.Contains(c.request.Header(acceptEncodingHeader), "gzip")

	if !c.b.gzip || !clientSupportsGZip || !canCompress(contentType) {
		header.Set(contentLengthHeader, strconv.Itoa(len(body)))
		c.response.inner.WriteHeader(c.status)
		_, err := c.response.inner.Write(body)
		return err
	}

	// GZip
	header.Set(contentEncodingHeader, contentEncodingGzip)
	c.response.inner.WriteHeader(c.status)

	// Write response body
	writer := c.b.acquireGZipWriter(c.response.inner)
	_, err := writer.Write(body)
	writer.Close()

	// Put the writer back into the pool
	c.b.gzipWriterPool.Put(writer)

	// Return the error value of the last Write call
	return err
}

// addParameter adds a new parameter to the context.
func (c *Context) addParameter(name string, value string) {
	c.paramNames[c.paramCount] = name
	c.paramValues[c.paramCount] = value
	c.paramCount++
}

// JSON encodes the object to a JSON string and responds.
func (c *Context) JSON(value interface{}) error {
	c.response.SetHeader(contentTypeHeader, contentTypeJSON)
	bytes, err := Json.Marshal(value)

	if err != nil {
		return err
	}

	return c.Bytes(bytes)
}

// HTML sends a HTML string.
func (c *Context) HTML(html string) error {
	header := c.response.inner.Header()
	header.Set(contentTypeHeader, contentTypeHTML)
	header.Set(contentTypeOptionsHeader, contentTypeOptions)
	header.Set(xssProtectionHeader, xssProtection)
	header.Set(referrerPolicyHeader, referrerPolicySameOrigin)

	return c.String(html)
}

// Close frees up resources and is automatically called
// in the ServeHTTP part of the web server.
func (c *Context) Close() {
	c.b.contextPool.Put(c)
}

// CSS sends a style sheet.
func (c *Context) CSS(text string) error {
	c.response.SetHeader(contentTypeHeader, contentTypeCSS)
	return c.String(text)
}

// JavaScript sends a script.
func (c *Context) JavaScript(code string) error {
	c.response.SetHeader(contentTypeHeader, contentTypeJavaScript)
	return c.String(code)
}

// EventStream sends server events to the client.
func (c *Context) EventStream(stream *EventStream) error {
	defer close(stream.Closed)

	// Flush
	flusher, ok := c.response.inner.(http.Flusher)

	if !ok {
		return c.Error(http.StatusNotImplemented, "Flushing not supported")
	}

	// Catch disconnect events
	disconnectedContext := c.request.Context()
	disconnectedContext, cancel := stdContext.WithDeadline(disconnectedContext, time.Now().Add(2*time.Hour))
	disconnected := disconnectedContext.Done()
	defer cancel()

	// Send headers
	header := c.response.inner.Header()
	header.Set(contentTypeHeader, contentTypeEventStream)
	header.Set(cacheControlHeader, "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("Access-Control-Allow-Origin", "*")
	c.response.inner.WriteHeader(200)

	for {
		select {
		case <-disconnected:
			return nil

		case event := <-stream.Events:
			if event != nil {
				data := event.Data

				switch data.(type) {
				case string, []byte:
					// Do nothing with the data if it's already a string or byte slice.
				default:
					var err error
					data, err = Json.Marshal(data)
					if err != nil {
						Log.Error(fmt.Sprintf("Failed encoding event data as JSON: %v", data))
					}
				}

				Log.Debug(fmt.Sprintf("event: %s\ndata: %s\n\n", event.Name, data))
				flusher.Flush()
			}
		}
	}
}

// File sends the contents of a local file and determines its mime type by extension.
func (c *Context) File(file string) error {
	extension := filepath.Ext(file)
	contentType := mime.TypeByExtension(extension)

	// Cache control header
	if isMedia(contentType) {
		c.response.SetHeader(cacheControlHeader, cacheControlMedia)
	}

	http.ServeFile(c.response.inner, c.request.inner, file)
	return nil
}

// Error should be used for sending error messages to the client.
func (c *Context) Error(statusCode int, errorList ...interface{}) error {
	c.status = statusCode

	if len(errorList) == 0 {
		message := http.StatusText(statusCode)
		_ = c.String(message)
		return errors.New(message)
	}

	messageBuffer := strings.Builder{}

	for index, param := range errorList {
		switch err := param.(type) {
		case string:
			messageBuffer.WriteString(err)
		case error:
			messageBuffer.WriteString(err.Error())
		default:
			continue
		}

		if index != len(errorList)-1 {
			messageBuffer.WriteString(": ")
		}
	}

	message := messageBuffer.String()
	_ = c.String(message)
	return errors.New(message)
}

// Path returns the relative request path, e.g. /blog/post/123.
func (c *Context) Path() string {
	return c.request.inner.URL.Path
}

// SetPath sets the relative request path, e.g. /blog/post/123.
func (c *Context) SetPath(path string) {
	c.request.inner.URL.Path = path
}

// Get retrieves an URL parameter.
func (c *Context) Get(param string) string {
	for i := 0; i < c.paramCount; i++ {
		if c.paramNames[i] == param {
			return c.paramValues[i]
		}
	}

	return ""
}

// GetInt retrieves an URL parameter as an integer.
func (c *Context) GetInt(param string) (int, error) {
	return strconv.Atoi(c.Get(param))
}

// Get IP by RemoteAddr
func (c *Context) IP() string {
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.request.inner.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

// ClientIP tries to determine the real IP address of the client.
func (c *Context) ClientIP() string {
	ip := c.request.Header(forwardedForHeader)
	ip = strings.TrimSpace(strings.Split(ip, ",")[0])
	if ip == "" {
		ip = strings.TrimSpace(c.request.Header(realIPHeader))
	}
	if ip != "" {
		return ip
	}

	if addr := c.request.Header(appengineRemoteAddr); addr != "" {
		return addr
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.request.inner.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

// Query retrieves the value for the given URL query parameter.
func (c *Context) Query(param string) string {
	return c.request.inner.URL.Query().Get(param)
}

// Redirect redirects to the given URL.
func (c *Context) Redirect(status int, u string) error {
	c.status = status
	c.response.SetHeader("Location", u)
	c.response.inner.WriteHeader(c.status)
	return nil
}

// isMedia returns whether the given content type is a media type.
func isMedia(contentType string) bool {
	switch {
	case strings.HasPrefix(contentType, "image/"):
		return true
	case strings.HasPrefix(contentType, "video/"):
		return true
	case strings.HasPrefix(contentType, "audio/"):
		return true
	default:
		return false
	}
}

// canCompress returns whether the given content type should be compressed via gzip.
func canCompress(contentType string) bool {
	switch {
	case strings.HasPrefix(contentType, "image/") && contentType != contentTypeSVG:
		return false
	case strings.HasPrefix(contentType, "video/"):
		return false
	case strings.HasPrefix(contentType, "audio/"):
		return false
	default:
		return true
	}
}

// ReadAll returns the contents of the reader.
// This will create an in-memory copy and calculate the E-Tag before sending the data.
// Compression will be applied if necessary.
func (c *Context) ReadAll(reader io.Reader) error {
	data, err := ioutil.ReadAll(reader)

	if err != nil {
		return err
	}

	return c.Bytes(data)
}

// Reader sends the contents of the io.Reader without creating an in-memory copy.
// E-Tags will not be generated for the content and compression will not be applied.
// Use this function if your reader contains huge amounts of data.
func (c *Context) Reader(reader io.Reader) error {
	_, err := io.Copy(c.response.inner, reader)
	return err
}

// ReadSeeker sends the contents of the io.ReadSeeker without creating an in-memory copy.
// E-Tags will not be generated for the content and compression will not be applied.
// Use this function if your reader contains huge amounts of data.
func (c *Context) ReadSeeker(reader io.ReadSeeker) error {
	http.ServeContent(c.response.inner, c.request.inner, "", time.Time{}, reader)
	return nil
}

// Status returns the HTTP status.
func (c *Context) Status() int {
	return c.status
}

// SetStatus sets the HTTP status.
func (c *Context) SetStatus(status int) {
	c.status = status
}

// String responds either with raw text or gzipped if the
// text length is greater than the gzip threshold.
func (c *Context) String(body string) error {
	return c.Bytes(bytesconv.StringToBytes(body))
}

// Request returns the HTTP request.
func (c *Context) Request() Request {
	return &c.request
}

// Response returns the HTTP response.
func (c *Context) Response() Response {
	return &c.response
}

// Text sends a plain text string.
func (c *Context) Text(text string) error {
	c.response.SetHeader(contentTypeHeader, contentTypePlainText)
	return c.String(text)
}

// ShouldBindJSON is a shortcut for c.ShouldBindWith(obj, binding.JSON).
func (c *Context) ShouldBindJSON(obj interface{}) error {
	return c.ShouldBindWith(obj, binding.JSON)
}

// ShouldBindWith binds the passed struct pointer using the specified binding engine.
// See the binding package.
func (c *Context) ShouldBindWith(obj interface{}, b binding.Binding) error {
	return b.Bind(c.request.Internal(), obj)
}

// Bind checks the Content-Type to select a binding engine automatically,
// Depending the "Content-Type" header different bindings are used:
//     "application/json" --> JSON binding
//     "application/xml"  --> XML binding
// otherwise --> returns an error.
// It parses the request's body as JSON if Content-Type == "application/json" using JSON or XML as a JSON input.
// It decodes the json payload into the struct specified as a pointer.
// It writes a 400 error and sets Content-Type header "text/plain" in the response if input is not valid.
func (c *Context) Bind(obj interface{}) error {
	b := binding.Default(c.Request().Method(), c.Request().Header(contentTypeHeader))
	return c.MustBindWith(obj, b)
}

// MustBindWith binds the passed struct pointer using the specified binding engine.
// It will abort the request with HTTP 400 if any error occurs.
// See the binding package.
func (c *Context) MustBindWith(obj interface{}, b binding.Binding) error {
	if err := c.ShouldBindWith(obj, b); err != nil {
		return c.Error(http.StatusBadRequest, err)
	}
	return nil
}

// Get name cookie value
func (c *Context) Cookie(name string) (string, error) {
	cookie, err := c.request.inner.Cookie(name)
	if err != nil {
		return "", err
	}
	v, _ := url.QueryUnescape(cookie.Value)
	return v, nil
}

// SetSameSite with cookie
func (c *Context) SetSameSite(samesite http.SameSite) {
	c.sameSite = samesite
}

// SetCookie adds a Set-Cookie header to the ResponseWriter's headers.
func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.response.inner, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		SameSite: c.sameSite,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}
