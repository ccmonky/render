package render

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ccmonky/errors"
	"github.com/ccmonky/inithook"
)

const (
	// ContentTypeHeader `Content-Type` header name
	ContentTypeHeader = "Content-Type"

	// AcceptHeader `Accept` header name
	AcceptHeader = "Accept"

	// TemplateHeader custom `X-Render-Template` header name, used to specify the render template,
	// can be used as request header and response header
	TemplateHeader = "X-Render-Template"
)

var (
	// Renders the renders registry, it's aim to store third party renders,
	// note, that's no need to store the `Response` render like JSON, HTML, ...
	Renders = inithook.NewMap[ContentType, Render]()

	// DefaultNegotiaterName default negotiater name
	DefaultNegotiaterName = ""

	// Negotiaters the negotiaters registry, now only the DefaultNegotiaterName used
	Negotiaters = inithook.NewMap[string, Negotiater]()
)

const (
	// JSON content type render for json
	JSON ContentType = "application/json; charset=utf-8"

	// JSONASCII content type render for json arcii
	JSONASCII ContentType = "application/json"

	// JSONP content type render for jsonp
	JSONP ContentType = "application/javascript; charset=utf-8"

	// HTML content type render for html
	HTML ContentType = "text/html; charset=utf-8"

	// Text content type render for text
	Text ContentType = "text/plain; charset=utf-8"

	PROTOBUF ContentType = "application/x-protobuf"

	// Binary content type render for binary
	Binary ContentType = "application/octet-stream"

	// YAML content type render for yaml
	YAML ContentType = "application/x-yaml; charset=utf-8"

	// TOML content type render for toml
	TOML ContentType = "application/toml; charset=utf-8"

	// MSGPACK content type render for msgpack
	MSGPACK ContentType = "application/msgpack; charset=utf-8"

	// XML content type render for xml
	XML ContentType = "application/xml; charset=utf-8"

	// XHTML content type render for xhtml
	XHTML ContentType = "application/xhtml+xml; charset=utf-8"
)

// Redner defines method to render any to http response writer
type Render interface {
	Render(http.ResponseWriter, interface{}) error
}

// RenderFunc defines the function that implement Render
type RenderFunc func(http.ResponseWriter, interface{}) error

func (rf RenderFunc) Render(w http.ResponseWriter, data interface{}) error {
	return rf(w, data)
}

// ContentType defines the content type render
type ContentType string

func (ct ContentType) OK() bool {
	return Renders.Has(context.Background(), ct)
}

// Header returns header value slice of content type
func (ct ContentType) Header() []string {
	return []string{string(ct)}
}

// Render implement `Render` interface, mainly used to extra suppport `ResponseInterface`
func (ct ContentType) Render(w http.ResponseWriter, rp interface{}) error {
	render, err := Renders.Get(context.TODO(), ct)
	if err != nil {
		return errors.WithMessagef(err, "get render failed for %v", ct)
	}
	switch rp := rp.(type) {
	case ResponseInterface:
		header := w.Header()
		if val := header[ContentTypeHeader]; len(val) == 0 {
			header[ContentTypeHeader] = ct.Header()
		}
		for k, vs := range rp.Header() {
			for _, v := range vs {
				header.Add(k, v)
			}
		}
		w.WriteHeader(rp.Status())
		return render.Render(w, rp.Body())
	default:
		return render.Render(w, rp)
	}
}

// Negotiate used to select the response content-type according to http request, default to `JSON`
// The default behavior can be changed, e.g.
//
//     import "github.com/timewasted/go-accept-headers"
//
//     type AcceptNegotiater struct{}
//
//     func (n AcceptNegotiater) Negotiate(acceptHeader string, ctypes ...string) (ctype string, err error) {
//     	   return accept.Parse(acceptHeader).Negotiate(ctypes...)
//     }
//
//     func init(){
//	       _ = render.Negotiaters.Set(ctx, render.DefaultNegotiaterName, AcceptNegotiater{})
//     }
//
func Negotiate(r *http.Request) ContentType {
	if r == nil {
		return defaultRender
	}
	header := r.Header.Get(AcceptHeader)
	if header == "" {
		return defaultRender
	}
	var ctypes []string
	for _, k := range Renders.Keys(r.Context()) {
		ctypes = append(ctypes, string(k))
	}
	negotiaterName := DefaultNegotiaterName
	negotiater, err := Negotiaters.Get(r.Context(), negotiaterName)
	if err != nil {
		log.Panicf("get negotiater %s failed: %v", negotiaterName, err)
	}
	ctype, err := negotiater.Negotiate(header, ctypes...)
	if err != nil {
		log.Panicf("negotiate failed: %v", err)
	}
	ct := ContentType(ctype)
	_, err = Renders.Get(r.Context(), ct)
	if err != nil {
		log.Panicf("render not found for %v", ct)
	}
	return ct
}

// Negotiater used to negotiate content type between client accepts and server supports
type Negotiater interface {
	Negotiate(acceptHeader string, ctypes ...string) (ctype string, err error)
}

type defaultNegotiater struct{}

func (n defaultNegotiater) Negotiate(acceptHeader string, ctypes ...string) (ctype string, err error) {
	return string(JSON), nil
}

// jsonRender implement `Render` for json format as default render
type jsonRender struct{}

// Render encode data as json bytes then write into the response writer
func (r jsonRender) Render(w http.ResponseWriter, data any) (err error) {
	header := w.Header()
	if val := header[ContentTypeHeader]; len(val) == 0 {
		header[ContentTypeHeader] = JSON.Header()
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write(bytes)
	return err
}

func init() {
	ctx := context.Background()
	Renders.Register(ctx, JSON, jsonRender{})
	Negotiaters.Register(ctx, DefaultNegotiaterName, defaultNegotiater{})
}

var (
	defaultRender = JSON
)

var (
	_ Render = (*ContentType)(nil)
	_ Render = (*RenderFunc)(nil)
)
