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

type JSON struct{}

func (r JSON) Render(w http.ResponseWriter, rp interface{}) error {
	return gin.JSON{Data: rp}.Render(w)
}

type IndentedJSON struct{}

func (r IndentedJSON) Render(w http.ResponseWriter, rp interface{}) error {
	return gin.IndentedJSON{Data: rp}.Render(w)
}

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

type AsciiJSON struct{}

func (r AsciiJSON) Render(w http.ResponseWriter, rp interface{}) error {
	return gin.AsciiJSON{Data: rp}.Render(w)
}

type PureJSON struct{}

func (r PureJSON) Render(w http.ResponseWriter, rp interface{}) error {
	return gin.PureJSON{Data: rp}.Render(w)
}

type Data struct{}

func (r Data) Render(w http.ResponseWriter, rp interface{}) error {
	if data, ok := rp.([]byte); ok {
		return gin.Data{
			ContentType: string(render.Binary),
			Data:        data,
		}.Render(w)
	}
	return fmt.Errorf("object of gin data render should be []byte, but got %T", rp)
}

type MsgPack struct{}

func (r MsgPack) Render(w http.ResponseWriter, rp interface{}) error {
	return gin.MsgPack{Data: rp}.Render(w)
}

type ProtoBuf struct{}

func (r ProtoBuf) Render(w http.ResponseWriter, rp interface{}) error {
	return gin.ProtoBuf{Data: rp}.Render(w)
}

// TODO
type String struct {
	Format string
	Data   []any
}

type TOML struct{}

func (r TOML) Render(w http.ResponseWriter, rp interface{}) error {
	return gin.TOML{Data: rp}.Render(w)
}

type XML struct{}

func (r XML) Render(w http.ResponseWriter, rp interface{}) error {
	return gin.XML{Data: rp}.Render(w)
}

type YAML struct{}

func (r YAML) Render(w http.ResponseWriter, rp interface{}) error {
	return gin.YAML{Data: rp}.Render(w)
}
