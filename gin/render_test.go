package gin_test

import (
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

type NoTimestampResponse struct {
	*render.Response
}

func (ntr NoTimestampResponse) Body() any {
	m := ntr.Response.Body().(map[string]any)
	delete(m, "timestamp")
	return m
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

func TestGin(t *testing.T) {

}
