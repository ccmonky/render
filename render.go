package render

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ccmonky/errors"
	"github.com/ccmonky/inithook"
	"github.com/timewasted/go-accept-headers"
)

const (
	ContentTypeHeader = "Content-Type"
	AcceptHeader      = "Accept"
	TemplateHeader    = "X-Render-Template"
)

const (
	// ContentJSON header value for JSON data.
	JSON ContentType = "application/json; charset=utf-8"

	JSONASCII ContentType = "application/json"

	// JSONP header value for JSONP data.
	JSONP ContentType = "application/javascript; charset=utf-8"

	// ContentHTML header value for HTML data.
	HTML ContentType = "text/html; charset=utf-8"

	// ContentText header value for Text data.
	Text ContentType = "text/plain; charset=utf-8"

	PROTOBUF ContentType = "application/x-protobuf"

	// Binary header value for binary data.
	Binary ContentType = "application/octet-stream"

	YAML ContentType = "application/x-yaml; charset=utf-8"

	TOML ContentType = "application/toml; charset=utf-8"

	MSGPACK ContentType = "application/msgpack; charset=utf-8"

	// ContentXML header value for XML data.
	XML ContentType = "application/xml; charset=utf-8"

	// ContentXHTML header value for XHTML data.
	XHTML ContentType = "application/xhtml+xml; charset=utf-8"
)

type Render interface {
	Render(http.ResponseWriter, interface{}) error
}

type RenderFunc func(http.ResponseWriter, interface{}) error

func (rf RenderFunc) Render(w http.ResponseWriter, data interface{}) error {
	return rf(w, data)
}

type ContentType string

func (ct ContentType) OK() bool {
	return Renders.Has(context.Background(), ct)
}

func (ct ContentType) Header() []string {
	return []string{string(ct)}
}

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

// GetRenderer according to request's `Accept` header
func Negotiate(r *http.Request) ContentType {
	if r == nil {
		return defaultRender
	}
	header := r.Header.Get(AcceptHeader)
	if header == "" {
		return defaultRender
	}
	acceptSlice := accept.Parse(header)
	var ctypes []string
	for _, k := range Renders.Keys(r.Context()) {
		ctypes = append(ctypes, string(k))
	}
	ctype, err := acceptSlice.Negotiate(ctypes...)
	if err != nil {
		log.Panicf("negotiate failed for request accept %s: %v", ctype, err)
	}
	ct := ContentType(ctype)
	_, err = Renders.Get(r.Context(), ct)
	if err != nil {
		log.Panicf("render not found for %v", ct)
	}
	return ct
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
	Renders.Register(context.Background(), JSON, jsonRender{})
}

var (
	defaultRender = JSON
	Renders       = inithook.NewMap[ContentType, Render]()
)

var (
	_ Render = (*ContentType)(nil)
	_ Render = (*RenderFunc)(nil)
)
