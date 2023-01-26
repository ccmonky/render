package render_test

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ccmonky/errors"
	"github.com/ccmonky/inithook"
	"github.com/ccmonky/render"
	"github.com/stretchr/testify/assert"
	"github.com/timewasted/go-accept-headers"
)

func init() {
	err := inithook.ExecuteAttrSetters(context.Background(), inithook.AppName, "myapp")
	if err != nil {
		panic(err)
	}
	err = inithook.ExecuteAttrSetters(context.Background(), inithook.Version, "0.3.0")
	if err != nil {
		panic(err)
	}
	err = render.Transformers.Register(context.Background(), "no_timestamp", func(rp *render.Response) render.ResponseInterface {
		return &NoTimestampResponse{rp}
	})
	if err != nil {
		panic(err)
	}
}

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

func TestResponseWriter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.JSON.Render(w, render.NewResponse(map[string]any{
			"one":    1,
			"string": "string",
		}, render.E(errors.NotFound), render.T("no_timestamp")))
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equalf(t, 404, res.StatusCode, "status")
	assert.Equalf(t, string(render.JSON), res.Header.Get("Content-Type"), "content-type")
	assert.Equalf(t, "myapp", res.Header.Get("X-App"), "x-app")
	assert.Equalf(t, "0.3.0", res.Header.Get("X-Version"), "X-Version")
	assert.Equalf(t, "not_found(5)", res.Header.Get("X-Code"), "x-code")
	assert.Equalf(t, "meta={source=errors;code=not_found(5)}:status={404}", res.Header.Get("X-Detail"), "x-detail")
	assert.Equalf(t, "not found", res.Header.Get("X-Message"), "x-message")
	assert.Equalf(t, "no_timestamp", res.Header.Get("X-Render-Template"), "X-Render-Template")
	expect := `{
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
	assert.JSONEq(t, expect, string(body), "body")
}

func TestOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.R("json").OK(w, r, map[string]any{
			"one":    1,
			"string": "string",
		})
	}))
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(render.TemplateHeader, "no_timestamp")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	assert.Equalf(t, 200, res.StatusCode, "status")
	assert.Equalf(t, string(render.JSON), res.Header.Get("Content-Type"), "content-type")
	assert.Equalf(t, "myapp", res.Header.Get("X-App"), "x-app")
	assert.Equalf(t, "0.3.0", res.Header.Get("X-Version"), "X-Version")
	assert.Equalf(t, "success(0)", res.Header.Get("X-Code"), "x-code")
	assert.Equalf(t, "meta={source=errors;code=success(0)}:status={200}", res.Header.Get("X-Detail"), "x-detail")
	assert.Equalf(t, "success", res.Header.Get("X-Message"), "x-message")
	assert.Equalf(t, "no_timestamp", res.Header.Get("X-Render-Template"), "X-Render-Template")
	expect := `{
		"app": "myapp",
		"code": "success(0)",
		"data": {
			"one": 1,
			"string": "string"
		},
		"detail": "meta={source=errors;code=success(0)}:status={200}",
		"message": "success",
		"version": "0.3.0"
	}`
	assert.JSONEq(t, expect, string(body), "body")
}

func TestErr(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.N(r).Err(w, r, errors.WithError(errors.New("xxx"), errors.AlreadyExists))
	}))
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(render.TemplateHeader, "no_timestamp")
	req.Header.Set(render.AcceptHeader, "application/json, text/html, application/xhtml+xml, application/xml;q=0.9, */*;q=0.8")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	assert.Equalf(t, 409, res.StatusCode, "status")
	assert.Equalf(t, string(render.JSON), res.Header.Get("Content-Type"), "content-type")
	assert.Equalf(t, "myapp", res.Header.Get("X-App"), "x-app")
	assert.Equalf(t, "0.3.0", res.Header.Get("X-Version"), "X-Version")
	assert.Equalf(t, "already_exists(6)", res.Header.Get("X-Code"), "x-code")
	assert.Equalf(t, "xxx:error={meta={source=errors;code=already_exists(6)}:status={409}}", res.Header.Get("X-Detail"), "x-detail")
	assert.Equalf(t, "already exists", res.Header.Get("X-Message"), "x-message")
	assert.Equalf(t, "no_timestamp", res.Header.Get("X-Render-Template"), "X-Render-Template")
	expect := `{
		"app": "myapp",
		"code": "already_exists(6)",
		"data": null,
		"detail": "xxx:error={meta={source=errors;code=already_exists(6)}:status={409}}",
		"message": "already exists",
		"version": "0.3.0"
	}`
	assert.JSONEq(t, expect, string(body), "body")
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
	assert.Equalf(t, string(render.JSON), ctype, "accept=%s", header)

	header = "application/json"
	ctype, err = n.Negotiate(header, string(render.JSON))
	assert.Nilf(t, err, "negotiate err")
	assert.Equalf(t, string(render.JSON), ctype, "accept=%s", header)

	header = "application/*"
	ctype, err = n.Negotiate(header, string(render.JSON))
	assert.Nilf(t, err, "negotiate err")
	assert.Equalf(t, string(render.JSON), ctype, "accept=%s", header)

	header = "text/html, application/xhtml+xml, application/xml;q=0.9, */*;q=0.8"
	ctype, err = n.Negotiate(header, string(render.JSON))
	assert.Nilf(t, err, "negotiate err")
	assert.Equalf(t, string(render.JSON), ctype, "accept=%s", header)

	ctype, err = n.Negotiate(header, string(render.JSON), string(render.XML))
	assert.Nilf(t, err, "negotiate err")
	assert.Equalf(t, string(render.XML), ctype, "accept=%s", header)

	ctype, err = n.Negotiate(header, string(render.JSON), string(render.XML), string(render.XHTML))
	assert.Nilf(t, err, "negotiate err")
	assert.Equalf(t, string(render.XHTML), ctype, "accept=%s", header)

	ctype, err = n.Negotiate(header, string(render.JSON), string(render.XML), string(render.XHTML), string(render.HTML))
	assert.Nilf(t, err, "negotiate err")
	assert.Equalf(t, string(render.HTML), ctype, "accept=%s", header)
}

type AcceptNegotiater struct{}

func (n AcceptNegotiater) Negotiate(acceptHeader string, ctypes ...string) (ctype string, err error) {
	return accept.Parse(acceptHeader).Negotiate(ctypes...)
}
