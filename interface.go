package render

import (
	"net/http"

	"github.com/ccmonky/inithook"
	"github.com/timewasted/go-accept-headers"
)

const (
	AcceptHeader = "Accept"
)

const (
	TemplateHeader = "X-Render-Template"
)

type ContentType = string

const (
	// ContentBinary header value for binary data.
	Binary ContentType = "application/octet-stream"

	// ContentHTML header value for HTML data.
	HTML ContentType = "text/html"

	// ContentJSON header value for JSON data.
	JSON ContentType = "application/json"

	// ContentJSONP header value for JSONP data.
	JSONP ContentType = "application/javascript"

	// ContentText header value for Text data.
	Text ContentType = "text/plain"

	// ContentXHTML header value for XHTML data.
	XHTML ContentType = "application/xhtml+xml"

	// ContentXML header value for XML data.
	XML ContentType = "text/xml"
)

// Render interface is to be implemented by JSON, XML, HTML, YAML and so on.
// provider: https://github.com/gin-gonic/gin/blob/master/render/render.go
// type Render interface {
// 	// Render writes data with custom ContentType.
// 	Render(http.ResponseWriter) error
// 	// WriteContentType writes custom ContentType.
// 	WriteContentType(w http.ResponseWriter)
// }

// Engine is the generic interface for all responses.
// provider: https://github.com/unrolled/render/blob/v1/engine.go
// type Engine interface {
// 	Render(io.Writer, interface{}) error
// }

type Render interface {
	Render(http.ResponseWriter, ResponseInterface) error
}

func NewRender(opts ...Option) Render {
	return nil
}

// Usage
// render.Use(r).Render(w)
func R(opts ...Option) Render {

	return nil
}

type Options struct {
	Format   string
	Template string
	*http.Request
	Extension map[string]any
}

type Option func(*Options)

func WithFormat(format string) Option {
	return func(options *Options) {
		options.Format = format
	}
}

func WithRequest(rq *http.Request) Option {
	return func(options *Options) {
		options.Request = rq
	}
}

func WithExtestion(k string, v any) Option {
	return func(options *Options) {
		options.Extension[k] = v
	}
}

// GetRenderer according to request's `Accept` header
func GetRenderer(r *http.Request) (Render, error) {
	if r == nil {
		return defaultRender, nil
	}
	header := r.Header.Get(AcceptHeader)
	acceptSlice := accept.Parse(header)
	ctype, err := acceptSlice.Negotiate(renders.Keys(r.Context())...)
	if err != nil {
		return nil, err
	}
	return renders.Get(r.Context(), ctype)
}

// func Render(w http.ResponseWriter, r *http.Request, response any) error {
// 	// 1. get renderer from request
// 	// 2. get template from request
// 	// 3. create response with err+data
// 	// 4. renderer.Render(w, response)
// 	return nil
// }

var (
	defaultRender Render
	renders       = inithook.NewMap[contentType, Render]()
)

type contentType = string
