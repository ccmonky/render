package gin

import (
	"context"
	"fmt"
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

func (sr SimpleRender[T]) Render(w http.ResponseWriter, rp interface{}) error {
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

func (r SecureJSON) Render(w http.ResponseWriter, rp interface{}) error {
	return gin.SecureJSON{Prefix: r.Prefix, Data: rp}.Render(w)
}

// JsonpJSON contains the given interface object its callback.
type JsonpJSON struct {
	Callback string
}

func (r JsonpJSON) Render(w http.ResponseWriter, rp interface{}) error {
	return gin.JsonpJSON{Callback: r.Callback, Data: rp}.Render(w)
}

type Data struct{}

func (r Data) Render(w http.ResponseWriter, rp interface{}) error {
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

// TODO
type String struct {
	Format string
	Data   []any
}
