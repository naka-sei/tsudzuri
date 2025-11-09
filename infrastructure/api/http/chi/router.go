package chiadapter

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"

	"github.com/go-chi/chi/v5"
	"github.com/naka-sei/tsudzuri/pkg/trace"
	perrcode "github.com/naka-sei/tsudzuri/presentation/http/errcode"
	presp "github.com/naka-sei/tsudzuri/presentation/http/response"
	prouter "github.com/naka-sei/tsudzuri/presentation/router"
)

type Router struct {
	r chi.Router
}

func New(r chi.Router) *Router { return &Router{r: r} }

// Route implements grouping under a path prefix.
func (c *Router) Route(pattern string, fn func(r prouter.Router)) {
	c.r.Route(pattern, func(r chi.Router) {
		fn(&Router{r: r})
	})
}

func (c *Router) Post(pattern string, handler any, opts ...prouter.Option) {
	c.r.Method(http.MethodPost, pattern, c.wrap(handler, http.MethodPost, opts...))
}

func (c *Router) Get(pattern string, handler any, opts ...prouter.Option) {
	c.r.Method(http.MethodGet, pattern, c.wrap(handler, http.MethodGet, opts...))
}

func (c *Router) Put(pattern string, handler any, opts ...prouter.Option) {
	c.r.Method(http.MethodPut, pattern, c.wrap(handler, http.MethodPut, opts...))
}

func (c *Router) Delete(pattern string, handler any, opts ...prouter.Option) {
	c.r.Method(http.MethodDelete, pattern, c.wrap(handler, http.MethodDelete, opts...))
}

func (c *Router) wrap(handler any, method string, opts ...prouter.Option) http.HandlerFunc {
	cfg := defaultConfigFor(method, opts...)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, end := trace.StartSpan(r.Context(), "infrastructure/http/chi.wrap")
		defer end()

		res, err := c.invoke(ctx, r, handler)
		if err != nil {
			writeError(w, r, err)
			return
		}
		if res == nil {
			w.WriteHeader(cfg.SuccessStatus)
			return
		}
		// If handler returns an EmptyResponse (by value or pointer), write only status
		switch res.(type) {
		case presp.EmptyResponse, *presp.EmptyResponse:
			w.WriteHeader(cfg.SuccessStatus)
			return
		}
		writeJSON(w, r, cfg.SuccessStatus, res)
	}
}

func (c *Router) invoke(ctx context.Context, r *http.Request, handler any) (any, error) {
	hv := reflect.ValueOf(handler)
	ht := hv.Type()
	if ht.Kind() != reflect.Func {
		return nil, errors.New("handler must be a function")
	}

	// Supported signatures:
	// 1) func(context.Context, Req) (Res, error)
	// 2) func(context.Context) (Res, error)
	if ht.NumOut() != 2 || !isErrorType(ht.Out(1)) {
		return nil, errors.New("handler must return (Res, error)")
	}

	var args []reflect.Value
	// First arg must be context.Context
	if ht.NumIn() == 0 || !isContextType(ht.In(0)) {
		return nil, errors.New("first parameter must be context.Context")
	}
	args = append(args, reflect.ValueOf(ctx))

	if ht.NumIn() == 2 {
		// Decode JSON body into request type
		reqPtr := reflect.New(ht.In(1))
		if err := decodeJSON(r, reqPtr.Interface()); err != nil {
			return nil, err
		}
		args = append(args, reqPtr.Elem())
	} else if ht.NumIn() > 2 {
		return nil, errors.New("handler must have 1 or 2 parameters (context, [req])")
	}

	outs := hv.Call(args)
	res := outs[0].Interface()
	var err error
	if !outs[1].IsNil() {
		err = outs[1].Interface().(error)
	}
	return res, err
}

// Utilities

func isContextType(t reflect.Type) bool {
	return t == reflect.TypeOf((*context.Context)(nil)).Elem()
}

func isErrorType(t reflect.Type) bool {
	return t == reflect.TypeOf((*error)(nil)).Elem()
}

func decodeJSON(r *http.Request, v any) error {
	if r.Body == nil {
		return io.EOF
	}
	defer func() {
		_ = r.Body.Close()
	}()
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(v); err != nil {
		if errors.Is(err, io.EOF) {
			return nil // empty body treated as zero-value
		}
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, _ *http.Request, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	sc := perrcode.GetStatusCode(err)
	re := perrcode.GetErrorReason(err)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(sc)
	_ = json.NewEncoder(w).Encode(re)
}

func defaultConfigFor(method string, opts ...prouter.Option) *prouter.RouteConfig {
	cfg := &prouter.RouteConfig{}
	switch method {
	case http.MethodPost:
		cfg.SuccessStatus = http.StatusCreated
	default:
		cfg.SuccessStatus = http.StatusOK
	}
	for _, o := range opts {
		o(cfg)
	}
	return cfg
}
