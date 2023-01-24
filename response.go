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

var (
	// E is abbr. of `WithError`
	E = WithError

	// KV is abbr. of `WithKV`
	KV = WithKV

	// T is abbr. of `WithTemplate`
	T = WithTemplate
)

// ResponseInterface defines http response interface
type ResponseInterface interface {
	// Status returns http status
	Status() int

	// Header returns http headers
	Header() http.Header

	// Body returns the http body
	Body() any
}

// Response aims to provide a widely applicable `ResponseInterface` implementation,
// it's made up the following elements:
//
// - Data: any object that will be rendered by concrete render, e.g. gin/render.JSON, unrolled/render.XML, ...
// - MetaError: contributes the response meta from the error(include all values it carries), use `WithError`(E) to specify
// - Extension: any kvs used to extend the `Response`, use `WithKV`(KV) to specify
// - Template: used to specify the variant of `Response`, use `WithTemplate`(T) to specify
//
// See tests for more details.
type Response struct {
	errors.MetaError
	Data      any
	Extension map[any]any
	Template  string

	m map[string]any // NOTE: errors.Map(MetaError)
}

// NewResponse creates a new *Response instance and returns it or it's variant as `ResponseInterface`
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

// ResponseOption `Response` creation option func
type ResponseOption func(*Response)

// WithError used to specify `MetaError` of `Response`
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

// WithKV used to specify `Extension` elements of `Response`
func WithKV(k, v any) ResponseOption {
	return func(rp *Response) {
		if rp.Extension == nil {
			rp.Extension = make(map[any]any)
		}
		rp.Extension[k] = v
	}
}

// WithTemplate used to specify the returned `Response` variant
func WithTemplate(tmpl string) ResponseOption {
	return func(rp *Response) {
		rp.Template = tmpl
	}
}

// Status implement `ResponseInterface` as standard
func (rp *Response) Status() int {
	return errors.StatusAttr.Get(rp.MetaError)
}

// Header implement `ResponseInterface` as standard
func (rp *Response) Header() http.Header {
	var header = make(http.Header, 6)
	header.Set(TemplateHeader, rp.Template)

	// configured values
	header.Set("X-App", appName.Load())
	header.Set("X-Version", appVersion.Load())

	// // error meta values
	header.Set("X-Code", rp.MetaError.Code())
	header.Set("X-Message", rp.MetaError.Message())
	header.Set("X-Detail", fmt.Sprint(rp.MetaError))
	return header

}

// Body implement `ResponseInterface` as standard
func (rp *Response) Body() any {
	return map[string]any{
		// configured values
		"app":     appName.Load(),
		"version": appVersion.Load(),

		// error meta values
		"code":    rp.MetaError.Code(),
		"message": rp.MetaError.Message(),
		"detail":  fmt.Sprint(rp.MetaError),

		// dynamic values
		"timestamp": time.Now().Unix(),

		// biz values
		"data": rp.Data,
	}
}

// Get used to get value specified by key from Response's Extension or error's values
// if found, return the value and true, otherwise return nil and false
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

// ResponseTransformer used to transform `*Response` to it's variant
type ResponseTransformer func(*Response) ResponseInterface

// Transformers defines the ResponseTransformer registry, key is the transformer name
var Transformers = inithook.NewMap[string, ResponseTransformer]()

func init() {
	inithook.RegisterAttrSetter(inithook.AppName, "render", func(ctx context.Context, value string) error {
		appName.Store(value)
		return nil
	})
	inithook.RegisterAttrSetter(inithook.Version, "render", func(ctx context.Context, value string) error {
		appVersion.Store(value)
		return nil
	})
	ctx := context.Background()
	Transformers.Register(ctx, "", selfResponseTransformer)
}

func selfResponseTransformer(rp *Response) ResponseInterface {
	return rp
}

var (
	appName    = atomic.NewString("")
	appVersion = atomic.NewString("")
)

var (
	_ ResponseInterface = (*Response)(nil)
)
