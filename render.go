package render

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ccmonky/inithook"
	"github.com/timewasted/go-accept-headers"
)

const (
	AcceptHeader   = "Accept"
	TemplateHeader = "X-Render-Template"
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
		return err
	}
	switch rp := rp.(type) {
	case ResponseInterface:
		w.WriteHeader(rp.Status())
		for k, vs := range rp.Header() {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
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
		log.Panic(err)
	}
	return ContentType(ctype)
}

// jsonRender implement `Render` for json format as default render
type jsonRender struct{}

// Render encode data as json bytes then write into the response writer
func (r jsonRender) Render(w http.ResponseWriter, data any) (err error) {
	writeContentType(w, JSON.Header())
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write(bytes)
	return err
}

// copy from gin/render
func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
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
)
