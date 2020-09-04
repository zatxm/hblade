package hblade

import "strconv"

// ETag produces a hash for the given slice of bytes.
func ETag(b []byte) string {
	return strconv.FormatUint(BytesHash(b), 16)
}

// ETagString produces a hash for the given string.
func ETagString(b string) string {
	return strconv.FormatUint(StringHash(b), 16)
}
