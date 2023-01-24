package unrolled

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/ccmonky/errors"
	"github.com/ccmonky/render"
	unrolled "github.com/unrolled/render"
)

func init() {
	ctx := context.Background()
	r := unrolled.New() // NOTE: default injection, use `renders.Renders.Set` to override!
	err := render.Renders.Set(ctx, render.JSON, JSON(r))
	err = errors.WithError(err, render.Renders.Set(ctx, render.Binary, Data(r)))
	err = errors.WithError(err, render.Renders.Set(ctx, render.Text, Text(r)))
	err = errors.WithError(err, render.Renders.Set(ctx, render.XML, XML(r)))
	if err != nil {
		log.Println(errors.GetAllErrors(err))
	}
}

func JSON(r *unrolled.Render) render.Render {
	return render.RenderFunc(func(w http.ResponseWriter, v interface{}) error {
		return r.JSON(w, statusPlaceholder, v) // NOTE: render has finished writing the status code so far!
	})
}

func Data(r *unrolled.Render) render.Render {
	return render.RenderFunc(func(w http.ResponseWriter, v interface{}) error {
		if data, ok := v.([]byte); ok {
			return r.Data(w, 0, data)
		}
		return r.Data(w, statusPlaceholder, []byte(fmt.Sprintf("%v", v)))
	})
}

func Text(r *unrolled.Render) render.Render {
	return render.RenderFunc(func(w http.ResponseWriter, v interface{}) error {
		if data, ok := v.(string); ok {
			return r.Text(w, statusPlaceholder, data)
		}
		return r.Text(w, statusPlaceholder, fmt.Sprintf("%v", v))
	})
}

func XML(r *unrolled.Render) render.Render {
	return render.RenderFunc(func(w http.ResponseWriter, v interface{}) error {
		return r.XML(w, statusPlaceholder, v) // NOTE: render has finished writing the status code so far!
	})
}

var (
	statusPlaceholder = 0
)
