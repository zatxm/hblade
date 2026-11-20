package hblade

import "net/textproto"

type flow int

const (
	flowStop flow = iota
	flowBegin
	flowNext

	cacheControlHeader            = "Cache-Control"
	cacheControlAlwaysValidate    = "must-revalidate"
	cacheControlMedia             = "public, max-age=13824000"
	contentTypeOptionsHeader      = "X-Content-Type-Options"
	contentTypeOptions            = "nosniff"
	xssProtectionHeader           = "X-XSS-Protection"
	xssProtection                 = "1; mode=block"
	etagHeader                    = "ETag"
	contentTypeHeader             = "Content-Type"
	contentTypeHTML               = "text/html; charset=utf-8"
	contentTypeCSS                = "text/css; charset=utf-8"
	contentTypeJavaScript         = "text/javascript; charset=utf-8"
	contentTypeJSON               = "application/json; charset=utf-8"
	contentTypePlainText          = "text/plain; charset=utf-8"
	contentTypeEventStream        = "text/event-stream; charset=utf-8"
	contentTypeSVG                = "image/svg+xml"
	contentEncodingHeader         = "Content-Encoding"
	contentEncodingGzip           = "gzip"
	acceptEncodingHeader          = "Accept-Encoding"
	contentLengthHeader           = "Content-Length"
	ifNoneMatchHeader             = "If-None-Match"
	referrerPolicyHeader          = "Referrer-Policy"
	referrerPolicySameOrigin      = "no-referrer"
	strictTransportSecurityHeader = "Strict-Transport-Security"
	strictTransportSecurity       = "max-age=31536000; includeSubDomains; preload"
	contentSecurityPolicyHeader   = "Content-Security-Policy"
	forwardedForHeader            = "X-Forwarded-For"
	realIPHeader                  = "X-Real-Ip"
	appengineRemoteAddr           = "X-Appengine-Remote-Addr"
)

const varyHeader = "Vary"

// Pre-built, canonical header values assigned directly to http.Header to avoid
// the per-call slice allocation of Header.Set. They must never be mutated.
var (
	contentTypeHTMLValue            = []string{contentTypeHTML}
	contentTypeCSSValue             = []string{contentTypeCSS}
	contentTypeJavaScriptValue      = []string{contentTypeJavaScript}
	contentTypeJSONValue            = []string{contentTypeJSON}
	contentTypePlainTextValue       = []string{contentTypePlainText}
	cacheControlMediaValue          = []string{cacheControlMedia}
	cacheControlAlwaysValidateValue = []string{cacheControlAlwaysValidate}
	contentEncodingGzipValue        = []string{contentEncodingGzip}
	contentTypeOptionsValue         = []string{contentTypeOptions}
	xssProtectionValue              = []string{xssProtection}
	referrerPolicySameOriginValue   = []string{referrerPolicySameOrigin}

	canonicalXSSProtectionHeader = textproto.CanonicalMIMEHeaderKey(xssProtectionHeader)
)
