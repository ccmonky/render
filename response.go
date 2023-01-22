package render

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ccmonky/errors"
	"github.com/ccmonky/inithook"
	"go.uber.org/atomic"
)

const (
	ResponserHeader = "X-Responser"
)

var (
	E  = WithError
	KV = WithKV
	T  = WithTemplate
)

type ResponseInterface interface {
	// Status returns http status
	Status() int

	// Header returns http headers
	Header() http.Header

	// Body returns the http body
	Body() any
}

type Response struct {
	errors.MetaError
	Data      any
	Extension map[any]any
	Template  string

	m map[string]any // NOTE: errors.Map(MetaError)
}

func NewResponse(data any, opts ...ResponseOption) ResponseInterface {
	rp := &Response{
		Data: data,
	}
	for _, opt := range opts {
		opt(rp)
	}
	if rp.MetaError == nil {
		rp.MetaError = errors.OK
	}
	if rp.Template != "" {
		transformer, err := Transformers.Get(context.Background(), rp.Template)
		if err != nil {
			log.Panicf("response transformer %s not found", rp.Template)
		}
		return transformer(rp)
	}
	return rp
}

type ResponseOption func(*Response)

func WithError(err error) ResponseOption {
	return func(rp *Response) {
		var me errors.MetaError
		if err == nil {
			me = errors.OK
		} else {
			if merr, ok := err.(errors.MetaError); ok {
				me = merr
			} else {
				me = errors.WithError(err, errors.Unknown).(errors.MetaError)
			}
		}
		rp.MetaError = me
	}
}

func WithKV(k, v any) ResponseOption {
	return func(rp *Response) {
		if rp.Extension == nil {
			rp.Extension = make(map[any]any)
		}
		rp.Extension[k] = v
	}
}

func WithTemplate(tmpl string) ResponseOption {
	return func(rp *Response) {
		rp.Template = tmpl
	}
}

func (rp *Response) Status() int {
	return errors.StatusAttr.Get(rp.MetaError)
}

func (rp *Response) Header() http.Header {
	var header = make(http.Header, 5)
	// configured values
	header.Set("X-App", appName.Load())
	header.Set("X-Version", appVersion.Load())

	// // error meta values
	header.Set("X-Code", rp.MetaError.Code())
	header.Set("X-Message", rp.MetaError.Message())
	header.Set("X-Detail", fmt.Sprint(rp.MetaError))
	return header

}

func (rp *Response) Body() any {
	return map[string]any{
		// configured values
		"app":     appName.Load(),
		"version": appVersion.Load(),

		// error meta values
		"code":    rp.MetaError.Code(),
		"message": rp.MetaError.Message(),
		"detail":  rp.MetaError,

		// dynamic values
		"timestamp": time.Now().Unix(),

		// biz values
		"data": rp.Data,
	}
}

func (rp *Response) Get(key any) (any, bool) {
	return Get(rp, key)
}

func Get(rp *Response, key any) (any, bool) {
	if value, ok := rp.Extension[key]; ok {
		return value, true
	}
	if sk, ok := key.(string); ok {
		if rp.m == nil {
			rp.m = errors.Map(rp.MetaError)
		}
		if value, ok := rp.m[sk]; ok {
			return value, true
		}
	}
	return nil, false
}

type Transformer func(*Response) ResponseInterface

var Transformers = inithook.NewMap[string, Transformer]()

func init() {
	inithook.RegisterAttrSetter(inithook.AppName, "render", func(ctx context.Context, value string) error {
		appName.Store(value)
		return nil
	})
	inithook.RegisterAttrSetter("app_verison", "render", func(ctx context.Context, value string) error {
		appVersion.Store(value)
		return nil
	})
	ctx := context.Background()
	Transformers.Register(ctx, "", defaultTransformer)
}

func defaultTransformer(rp *Response) ResponseInterface {
	return rp
}

var (
	appName    = atomic.NewString("")
	appVersion = atomic.NewString("")
)

var (
	_ ResponseInterface = (*Response)(nil)
)
