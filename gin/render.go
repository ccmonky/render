package gin

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/ccmonky/errors"
	"github.com/ccmonky/render"
	gin "github.com/gin-gonic/gin/render"
)

func init() {
	ctx := context.Background()
	err := render.Renders.Set(ctx, render.JSON, JSON{})
	err = errors.WithError(err, render.Renders.Register(ctx, render.Binary, Data{}))
	err = errors.WithError(err, render.Renders.Register(ctx, render.MSGPACK, MsgPack{}))
	err = errors.WithError(err, render.Renders.Register(ctx, render.PROTOBUF, ProtoBuf{}))
	err = errors.WithError(err, render.Renders.Register(ctx, render.TOML, TOML{}))
	err = errors.WithError(err, render.Renders.Register(ctx, render.XML, XML{}))
	err = errors.WithError(err, render.Renders.Register(ctx, render.YAML, YAML{}))
	err = errors.WithError(err, render.Renders.Register(ctx, render.Text, String{}))
	err = errors.WithError(err, render.Renders.Register(ctx, render.JSONP, JsonpJSON{}))
	err = errors.WithError(err, render.Renders.Register(ctx, render.HTML, HTML{}))
	if err != nil {
		log.Println(errors.GetAllErrors(err))
	}
}

type SimpleRenderKind interface {
	~struct {
		Data any
	}
	Render(http.ResponseWriter) error
}

type SimpleRender[T SimpleRenderKind] struct{}

func (sr SimpleRender[T]) Render(w http.ResponseWriter, rp any, opts ...render.Option) error {
	return T{Data: rp}.Render(w)
}

type JSON = SimpleRender[gin.JSON]
type IndentedJSON = SimpleRender[gin.IndentedJSON]
type AsciiJSON = SimpleRender[gin.AsciiJSON]
type PureJSON = SimpleRender[gin.PureJSON]
type MsgPack = SimpleRender[gin.MsgPack]
type ProtoBuf = SimpleRender[gin.ProtoBuf]
type TOML = SimpleRender[gin.TOML]
type XML = SimpleRender[gin.XML]
type YAML = SimpleRender[gin.YAML]

type SecureJSON struct {
	Prefix string
}

func (r SecureJSON) Render(w http.ResponseWriter, rp any, opts ...render.Option) error {
	return gin.SecureJSON{Prefix: r.Prefix, Data: rp}.Render(w)
}

type Data struct{}

func (r Data) Render(w http.ResponseWriter, rp any, opts ...render.Option) error {
	if data, ok := rp.([]byte); ok {
		return gin.Data{
			ContentType: string(render.Binary),
			Data:        data,
		}.Render(w)
	}
	return gin.Data{
		ContentType: string(render.Binary),
		Data:        []byte(fmt.Sprintf("%v", rp)),
	}.Render(w)
}

// JsonpJSON contains the given interface object its callback.
type JsonpJSON struct{}

func (r JsonpJSON) Render(w http.ResponseWriter, rp any, opts ...render.Option) error {
	options := jsonpOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	return gin.JsonpJSON{Callback: options.Callback, Data: rp}.Render(w)
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

type String struct{}

func (r String) Render(w http.ResponseWriter, rp any, opts ...render.Option) error {
	options := stringOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	if args, ok := rp.([]any); ok {
		return gin.String{Format: options.Format, Data: args}.Render(w)
	}
	return fmt.Errorf("gin string data should be []any, but got %T", rp)
}

type stringOptions struct {
	Format string
}

func StringFormat(format string) render.Option {
	return func(o any) {
		if options, ok := o.(*stringOptions); ok {
			options.Format = format
		}
	}
}

type HTML struct{}

func (r HTML) Render(w http.ResponseWriter, rp any, opts ...render.Option) error {
	options := htmlOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	html := gin.HTML{
		Template: options.Template,
		Name:     options.Name,
		Data:     rp,
	}
	return html.Render(w)
}

type htmlOptions struct {
	Template *template.Template
	Name     string
}

// HTMLTemplate specify html tempalte
func HTMLTemplate(t *template.Template) render.Option {
	return func(o any) {
		if options, ok := o.(*htmlOptions); ok {
			options.Template = t
		}
	}
}

// HTMLName specify html template name
func HTMLName(name string) render.Option {
	return func(o any) {
		if options, ok := o.(*htmlOptions); ok {
			options.Name = name
		}
	}
}
