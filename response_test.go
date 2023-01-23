package render_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/ccmonky/inithook"
	"github.com/ccmonky/render"
	"github.com/stretchr/testify/assert"
)

func init() {
	err := inithook.ExecuteAttrSetters(context.Background(), inithook.AppName, "myapp")
	if err != nil {
		panic(err)
	}
	err = inithook.ExecuteAttrSetters(context.Background(), "app_version", "0.3.0")
	if err != nil {
		panic(err)
	}
}

type ReexportResponse render.Response

func (rr ReexportResponse) Status() int {
	return 404
}
func (rr ReexportResponse) Header() http.Header {
	return nil
}
func (rr ReexportResponse) Body() any {
	return nil
}

type EmbedResponse struct {
	*render.Response
}

func (er EmbedResponse) Status() int {
	return 500
}

func (er EmbedResponse) Header() http.Header {
	header := er.Response.Header()
	header.Set("embed", "true")
	return header
}

func TestResponse(t *testing.T) {
	data := map[string]any{
		"1": "one",
		"2": "two",
	}
	rp := render.NewResponse(data)
	assert.Equalf(t, 200, rp.Status(), "rp status")
	assert.Equalf(t, 6, len(rp.Header()), "rp header length")
	assert.Equalf(t, 7, len(rp.Body().(map[string]any)), "rp body length")

	ctx := context.Background()
	render.Transformers.Register(ctx, "reexport", func(rp *render.Response) render.ResponseInterface {
		return (*ReexportResponse)(rp)
	})
	render.Transformers.Register(ctx, "embed", func(rp *render.Response) render.ResponseInterface {
		return &EmbedResponse{rp}
	})

	rp = render.NewResponse(data, render.WithTemplate("reexport"))
	assert.Equalf(t, 404, rp.Status(), "rp reexport status")
	assert.Equalf(t, 0, len(rp.Header()), "rp reexport header length")
	assert.Equalf(t, nil, rp.Body(), "rp reexport body length")

	rp = render.NewResponse(data, render.WithTemplate("embed"))
	assert.Equalf(t, 500, rp.Status(), "rp embed status")
	assert.Equalf(t, 7, len(rp.Header()), "rp embed header length")
	assert.Equalf(t, 7, len(rp.Body().(map[string]any)), "rp embed body length")
}
