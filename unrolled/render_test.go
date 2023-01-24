package unrolled_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ccmonky/errors"
	"github.com/ccmonky/inithook"
	"github.com/ccmonky/render"
	"github.com/stretchr/testify/assert"
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
	if rw.status >= 200 {
		return
	}
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

func TestUnrolled(t *testing.T) {
	w := newResponseWriter()
	render.JSON.Render(w, nil)
	assert.Equalf(t, 200, w.status, "status")
	assert.Equalf(t, 1, len(w.header), "header")
	assert.Containsf(t, w.header, "Content-Type", "content-type header")
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

func TestWriteHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.WriteHeader(500)
		fmt.Fprint(w, "Hello, client")
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Error("should ==")
	}
	greeting, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	if string(greeting) != "Hello, client" {
		t.Fatalf("should ==, got %s", string(greeting))
	}
}
