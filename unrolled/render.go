package unrolled

import (
	"context"
	"log"
	"net/http"

	"github.com/ccmonky/errors"
	"github.com/ccmonky/render"
	unrolled "github.com/unrolled/render"
)

func init() {
	ctx := context.Background()
	r := unrolled.New()
	err := render.Renders.Set(ctx, render.JSON, JSON(r))
	if err != nil {
		log.Println(errors.GetAllErrors(err))
	}
}

func JSON(r *unrolled.Render) render.Render {
	return render.RenderFunc(func(w http.ResponseWriter, v interface{}) error {
		return r.JSON(w, 0, v) // NOTE: render has finished writing the status code so far!
	})
}
