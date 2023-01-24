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
	err = errors.WithError(err, render.Renders.Set(ctx, render.JSONP, JSONP(r)))
	err = errors.WithError(err, render.Renders.Set(ctx, render.HTML, HTML(r)))
	if err != nil {
		log.Println(errors.GetAllErrors(err))
	}
}

func JSON(r *unrolled.Render) render.Render {
	return render.RenderFunc(func(w http.ResponseWriter, v interface{}, opts ...render.Option) error {
		return r.JSON(w, statusPlaceholder, v) // NOTE: render has finished writing the status code so far!
	})
}

func Data(r *unrolled.Render) render.Render {
	return render.RenderFunc(func(w http.ResponseWriter, v interface{}, opts ...render.Option) error {
		if data, ok := v.([]byte); ok {
			return r.Data(w, 0, data)
		}
		return r.Data(w, statusPlaceholder, []byte(fmt.Sprintf("%v", v)))
	})
}

func Text(r *unrolled.Render) render.Render {
	return render.RenderFunc(func(w http.ResponseWriter, v interface{}, opts ...render.Option) error {
		if data, ok := v.(string); ok {
			return r.Text(w, statusPlaceholder, data)
		}
		return r.Text(w, statusPlaceholder, fmt.Sprintf("%v", v))
	})
}

func XML(r *unrolled.Render) render.Render {
	return render.RenderFunc(func(w http.ResponseWriter, v interface{}, opts ...render.Option) error {
		return r.XML(w, statusPlaceholder, v)
	})
}

func JSONP(r *unrolled.Render) render.Render {
	return render.RenderFunc(func(w http.ResponseWriter, v interface{}, opts ...render.Option) error {
		options := jsonpOptions{}
		for _, opt := range opts {
			opt(&options)
		}
		return r.JSONP(w, statusPlaceholder, options.Callback, v)
	})
}

type jsonpOptions struct {
	Callback string
}

func JSONPCallback(callback string) render.Option {
	return func(o any) {
		if options, ok := o.(*jsonpOptions); ok {
			options.Callback = callback
		}
	}
}

func HTML(r *unrolled.Render) render.Render {
	return render.RenderFunc(func(w http.ResponseWriter, v interface{}, opts ...render.Option) error {
		options := htmlOptions{}
		for _, opt := range opts {
			opt(&options)
		}
		return r.HTML(w, statusPlaceholder, options.Template, v, options.Options...)
	})
}

type htmlOptions struct {
	Template string
	Options  []unrolled.HTMLOptions
}

func HTMLTemplate(template string) render.Option {
	return func(o any) {
		if options, ok := o.(*htmlOptions); ok {
			options.Template = template
		}
	}
}

func HTMLOptions(opts ...unrolled.HTMLOptions) render.Option {
	return func(o any) {
		if options, ok := o.(*htmlOptions); ok {
			options.Options = opts
		}
	}
}

var (
	statusPlaceholder = 0
)
