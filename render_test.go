package render_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/ccmonky/errors"
	"github.com/ccmonky/render"
	"github.com/stretchr/testify/assert"
	"github.com/timewasted/go-accept-headers"
)

type responseWriter struct {
	status int
	header http.Header
	body   bytes.Buffer
}

func newResponseWriter() *responseWriter {
	return &responseWriter{
		header: make(http.Header),
	}
}

func (rw *responseWriter) Header() http.Header {
	return rw.header
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if rw.status == 0 {
		rw.WriteHeader(200)
	}
	return rw.body.Write(data)
}

type NoTimestampResponse struct {
	*render.Response
}

func (ntr NoTimestampResponse) Body() any {
	m := ntr.Response.Body().(map[string]any)
	delete(m, "timestamp")
	return m
}

func TestDefault(t *testing.T) {
	w := newResponseWriter()
	render.JSON.Render(w, nil)
	assert.Equalf(t, 200, w.status, "status")
	assert.Equalf(t, 1, len(w.header), "header")
	assert.Equalf(t, "null", w.body.String(), "body")

	w = newResponseWriter()
	render.JSON.Render(w, 101)
	assert.Equalf(t, 200, w.status, "status")
	assert.Equalf(t, 1, len(w.header), "header")
	assert.Equalf(t, "101", w.body.String(), "body")

	w = newResponseWriter()
	render.JSON.Render(w, map[string]any{
		"one":    1,
		"string": "string",
	})
	assert.Equalf(t, 200, w.status, "status")
	assert.Equalf(t, 1, len(w.header), "header")
	assert.JSONEq(t, `{"one":1,"string":"string"}`, w.body.String(), "body")

	w = newResponseWriter()
	render.Transformers.Register(context.Background(), "no_timestamp", func(rp *render.Response) render.ResponseInterface {
		return &NoTimestampResponse{rp}
	})
	render.JSON.Render(w, render.NewResponse(map[string]any{
		"one":    1,
		"string": "string",
	}, render.E(errors.NotFound), render.T("no_timestamp")))
	assert.Equalf(t, 404, w.status, "status")
	assert.Equalf(t, 7, len(w.header), "header")
	body := `{
		"app": "myapp",
		"code": "not_found(5)",
		"data": {
			"one": 1,
			"string": "string"
		},
		"detail": "meta={source=errors;code=not_found(5)}:status={404}",
		"message": "not found",
		"version": "0.3.0"
	}`
	assert.JSONEq(t, body, w.body.String(), "body")

	w = newResponseWriter()
	assert.NotNilf(t, render.XML.Render(w, nil), "default only json")
}

func TestNegotiate(t *testing.T) {
	ctx := context.Background()
	err := render.Negotiaters.Set(ctx, "", AcceptNegotiater{})
	assert.Nilf(t, err, "set negotiater")
	n, err := render.Negotiaters.Get(ctx, "")
	assert.Nilf(t, err, "get negotiater")

	header := ""
	ctype, err := n.Negotiate(header, string(render.JSON))
	assert.Nilf(t, err, "negotiate err")
	assert.Equalf(t, "application/json; charset=utf-8", ctype, "accept=%s", header)

	header = "application/json"
	ctype, err = n.Negotiate(header, string(render.JSON))
	assert.Nilf(t, err, "negotiate err")
	assert.Equalf(t, "application/json; charset=utf-8", ctype, "accept=%s", header)

	header = "application/*"
	ctype, err = n.Negotiate(header, string(render.JSON))
	assert.Nilf(t, err, "negotiate err")
	assert.Equalf(t, "application/json; charset=utf-8", ctype, "accept=%s", header)

	header = "text/html, application/xhtml+xml, application/xml;q=0.9, */*;q=0.8"
	ctype, err = n.Negotiate(header, string(render.JSON))
	assert.Nilf(t, err, "negotiate err")
	assert.Equalf(t, "application/json; charset=utf-8", ctype, "accept=%s", header)

	ctype, err = n.Negotiate(header, string(render.JSON), string(render.XML))
	assert.Nilf(t, err, "negotiate err")
	assert.Equalf(t, "application/xml; charset=utf-8", ctype, "accept=%s", header)

	ctype, err = n.Negotiate(header, string(render.JSON), string(render.XML), string(render.XHTML))
	assert.Nilf(t, err, "negotiate err")
	assert.Equalf(t, "application/xhtml+xml; charset=utf-8", ctype, "accept=%s", header)

	ctype, err = n.Negotiate(header, string(render.JSON), string(render.XML), string(render.XHTML), string(render.HTML))
	assert.Nilf(t, err, "negotiate err")
	assert.Equalf(t, "text/html; charset=utf-8", ctype, "accept=%s", header)
}

type AcceptNegotiater struct{}

func (n AcceptNegotiater) Negotiate(acceptHeader string, ctypes ...string) (ctype string, err error) {
	return accept.Parse(acceptHeader).Negotiate(ctypes...)
}
