package chiadapter

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	cmp "github.com/google/go-cmp/cmp"
	prouter "github.com/naka-sei/tsudzuri/presentation/router"
)

func TestNew(t *testing.T) {
	r := chi.NewRouter()
	router := New(r)
	if router == nil {
		t.Fatal("New returned nil")
	}
	if router.r != r {
		t.Error("Router's internal router not set correctly")
	}
}

func TestRoute(t *testing.T) {
	r := chi.NewRouter()
	router := New(r)

	called := false
	router.Route("/api", func(sub prouter.Router) {
		sub.Get("/test", func(ctx context.Context) (string, error) {
			called = true
			return "ok", nil
		})
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !called {
		t.Error("Handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	var resp string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}
	if resp != "ok" {
		t.Errorf("Expected 'ok', got %s", resp)
	}
}

func TestPost(t *testing.T) {
	type TestRequest struct {
		Name string `json:"name"`
	}
	type TestResponse struct {
		Message string `json:"message"`
	}

	type args struct {
		req *http.Request
	}
	type want struct {
		status int
		body   string
	}

	tests := []struct {
		name    string
		handler any
		args    args
		want    want
	}{
		{
			name: "success_with_request_body",
			handler: func(ctx context.Context, req TestRequest) (TestResponse, error) {
				return TestResponse{Message: "Hello " + req.Name}, nil
			},
			args: args{
				req: func() *http.Request {
					req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"world"}`))
					req.Header.Set("Content-Type", "application/json")
					return req
				}(),
			},
			want: want{
				status: http.StatusCreated,
				body:   `{"message":"Hello world"}`,
			},
		},
		{
			name: "success_with_request_body_pointer",
			handler: func(ctx context.Context, req *TestRequest) (TestResponse, error) {
				return TestResponse{Message: "Hello " + req.Name}, nil
			},
			args: args{
				req: func() *http.Request {
					req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"pointer"}`))
					req.Header.Set("Content-Type", "application/json")
					return req
				}(),
			},
			want: want{
				status: http.StatusCreated,
				body:   `{"message":"Hello pointer"}`,
			},
		},
		{
			name: "error",
			handler: func(ctx context.Context, req TestRequest) (TestResponse, error) {
				return TestResponse{}, errors.New("handler error")
			},
			args: args{
				req: func() *http.Request {
					req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"fail"}`))
					req.Header.Set("Content-Type", "application/json")
					return req
				}(),
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   `{"message":"不明なエラーが発生しました。再度お試しください。"}`,
			},
		},
		{
			name: "empty_response",
			handler: func(ctx context.Context) error {
				return nil
			},
			args: args{
				req: func() *http.Request {
					req := httptest.NewRequest("POST", "/test", strings.NewReader(""))
					req.Header.Set("Content-Type", "application/json")
					return req
				}(),
			},
			want: want{
				status: http.StatusCreated,
				body:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			router := New(r)
			router.Post("/test", tt.handler)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, tt.args.req)

			if w.Code != tt.want.status {
				t.Errorf("Expected status %d, got %d", tt.want.status, w.Code)
			}
			body := strings.TrimSpace(w.Body.String())
			if body != tt.want.body {
				t.Errorf("Expected body %q, got %q", tt.want.body, body)
			}
		})
	}
}

func TestGet(t *testing.T) {
	type args struct {
		req *http.Request
	}
	type want struct {
		status int
		body   string
	}

	tests := []struct {
		name    string
		handler any
		args    args
		want    want
	}{
		{
			name: "success",
			handler: func(ctx context.Context) (string, error) {
				return "get ok", nil
			},
			args: args{
				req: httptest.NewRequest("GET", "/test", nil),
			},
			want: want{
				status: http.StatusOK,
				body:   `"get ok"`,
			},
		},
		{
			name: "error",
			handler: func(ctx context.Context) (string, error) {
				return "", errors.New("handler error")
			},
			args: args{
				req: httptest.NewRequest("GET", "/test", nil),
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   `{"message":"不明なエラーが発生しました。再度お試しください。"}`,
			},
		},
		{
			name: "empty_response",
			handler: func(ctx context.Context) error {
				return nil
			},
			args: args{
				req: httptest.NewRequest("GET", "/test", nil),
			},
			want: want{
				status: http.StatusOK,
				body:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			router := New(r)
			router.Get("/test", tt.handler)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, tt.args.req)

			if w.Code != tt.want.status {
				t.Errorf("Expected status %d, got %d", tt.want.status, w.Code)
			}
			body := strings.TrimSpace(w.Body.String())
			if body != tt.want.body {
				t.Errorf("Expected body %q, got %q", tt.want.body, body)
			}
		})
	}
}

func TestPut(t *testing.T) {
	type args struct {
		req *http.Request
	}
	type want struct {
		status int
		body   string
	}

	tests := []struct {
		name    string
		handler any
		args    args
		want    want
	}{
		{
			name: "success",
			handler: func(ctx context.Context) (string, error) {
				return "put ok", nil
			},
			args: args{
				req: func() *http.Request {
					req := httptest.NewRequest("PUT", "/test", strings.NewReader(""))
					req.Header.Set("Content-Type", "application/json")
					return req
				}(),
			},
			want: want{
				status: http.StatusOK,
				body:   `"put ok"`,
			},
		},
		{
			name: "error",
			handler: func(ctx context.Context) (string, error) {
				return "", errors.New("handler error")
			},
			args: args{
				req: func() *http.Request {
					req := httptest.NewRequest("PUT", "/test", strings.NewReader(""))
					req.Header.Set("Content-Type", "application/json")
					return req
				}(),
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   `{"message":"不明なエラーが発生しました。再度お試しください。"}`,
			},
		},
		{
			name: "empty_response",
			handler: func(ctx context.Context) error {
				return nil
			},
			args: args{
				req: func() *http.Request {
					req := httptest.NewRequest("PUT", "/test", strings.NewReader(""))
					req.Header.Set("Content-Type", "application/json")
					return req
				}(),
			},
			want: want{
				status: http.StatusOK,
				body:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			router := New(r)
			router.Put("/test", tt.handler)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, tt.args.req)

			if w.Code != tt.want.status {
				t.Errorf("Expected status %d, got %d", tt.want.status, w.Code)
			}
			body := strings.TrimSpace(w.Body.String())
			if body != tt.want.body {
				t.Errorf("Expected body %q, got %q", tt.want.body, body)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		req *http.Request
	}
	type want struct {
		status int
		body   string
	}

	tests := []struct {
		name    string
		handler any
		args    args
		want    want
	}{
		{
			name: "success",
			handler: func(ctx context.Context) (string, error) {
				return "deleted", nil
			},
			args: args{
				req: httptest.NewRequest("DELETE", "/test", nil),
			},
			want: want{
				status: http.StatusOK,
				body:   `"deleted"`,
			},
		},
		{
			name: "error",
			handler: func(ctx context.Context) (string, error) {
				return "", errors.New("handler error")
			},
			args: args{
				req: httptest.NewRequest("DELETE", "/test", nil),
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   `{"message":"不明なエラーが発生しました。再度お試しください。"}`,
			},
		},
		{
			name: "empty_response",
			handler: func(ctx context.Context) error {
				return nil
			},
			args: args{
				req: httptest.NewRequest("DELETE", "/test", nil),
			},
			want: want{
				status: http.StatusOK,
				body:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			router := New(r)
			router.Delete("/test", tt.handler)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, tt.args.req)

			if w.Code != tt.want.status {
				t.Errorf("Expected status %d, got %d", tt.want.status, w.Code)
			}
			body := strings.TrimSpace(w.Body.String())
			if body != tt.want.body {
				t.Errorf("Expected body %q, got %q", tt.want.body, body)
			}
		})
	}
}

func TestInvoke(t *testing.T) {
	type TestReq struct {
		Name string `json:"name" path:"id"`
	}
	type TestRes struct {
		Greeting string `json:"greeting"`
	}

	type args struct {
		req *http.Request
	}
	type want struct {
		res   any
		error bool
	}

	tests := []struct {
		name    string
		handler any
		args    args
		want    want
	}{
		{
			name: "success_with_req",
			handler: func(ctx context.Context, req TestReq) (TestRes, error) {
				return TestRes{Greeting: "Hello " + req.Name}, nil
			},
			args: args{
				req: func() *http.Request {
					req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"Alice"}`))
					req.Header.Set("Content-Type", "application/json")
					return req
				}(),
			},
			want: want{
				res:   TestRes{Greeting: "Hello Alice"},
				error: false,
			},
		},
		{
			name: "success_with_req_pointer",
			handler: func(ctx context.Context, req *TestReq) (TestRes, error) {
				return TestRes{Greeting: "Hello " + req.Name}, nil
			},
			args: args{
				req: func() *http.Request {
					req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"Bob"}`))
					req.Header.Set("Content-Type", "application/json")
					return req
				}(),
			},
			want: want{
				res:   TestRes{Greeting: "Hello Bob"},
				error: false,
			},
		},
		{
			name: "success_without_req",
			handler: func(ctx context.Context) (TestRes, error) {
				return TestRes{Greeting: "Hello World"}, nil
			},
			args: args{
				req: func() *http.Request {
					req := httptest.NewRequest("POST", "/test", strings.NewReader(""))
					req.Header.Set("Content-Type", "application/json")
					return req
				}(),
			},
			want: want{
				res:   TestRes{Greeting: "Hello World"},
				error: false,
			},
		},
		{
			name: "with_path_param",
			handler: func(ctx context.Context, req TestReq) (TestRes, error) {
				return TestRes{Greeting: "ID: " + req.Name}, nil
			},
			args: args{
				req: func() *http.Request {
					req := httptest.NewRequest("POST", "/test/123", strings.NewReader(`{}`))
					req.Header.Set("Content-Type", "application/json")
					rctx := chi.NewRouteContext()
					rctx.URLParams.Add("id", "123")
					req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
					return req
				}(),
			},
			want: want{
				res:   TestRes{Greeting: "ID: 123"},
				error: false,
			},
		},
		{
			name: "with_path_param_pointer",
			handler: func(ctx context.Context, req *TestReq) (TestRes, error) {
				return TestRes{Greeting: "ID: " + req.Name}, nil
			},
			args: args{
				req: func() *http.Request {
					req := httptest.NewRequest("POST", "/test/456", strings.NewReader(`{}`))
					req.Header.Set("Content-Type", "application/json")
					rctx := chi.NewRouteContext()
					rctx.URLParams.Add("id", "456")
					req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
					return req
				}(),
			},
			want: want{
				res:   TestRes{Greeting: "ID: 456"},
				error: false,
			},
		},
		{
			name:    "invalid_handler_signature_no_params",
			handler: func() (TestRes, error) { return TestRes{}, nil },
			args: args{
				req: func() *http.Request {
					req := httptest.NewRequest("POST", "/test", strings.NewReader(""))
					req.Header.Set("Content-Type", "application/json")
					return req
				}(),
			},
			want: want{
				res:   nil,
				error: true,
			},
		},
		{
			name:    "invalid_handler_signature_wrong_return",
			handler: func(ctx context.Context) TestRes { return TestRes{} },
			args: args{
				req: func() *http.Request {
					req := httptest.NewRequest("POST", "/test", strings.NewReader(""))
					req.Header.Set("Content-Type", "application/json")
					return req
				}(),
			},
			want: want{
				res:   nil,
				error: true,
			},
		},
		{
			name:    "json_decode_error",
			handler: func(ctx context.Context, req TestReq) (TestRes, error) { return TestRes{}, nil },
			args: args{
				req: func() *http.Request {
					req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":}`))
					req.Header.Set("Content-Type", "application/json")
					return req
				}(),
			},
			want: want{
				res:   nil,
				error: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			router := New(r)

			res, err := router.invoke(tt.args.req.Context(), tt.args.req, tt.handler)

			if tt.want.error {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if diff := cmp.Diff(tt.want.res, res); diff != "" {
					t.Errorf("Response mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestIsContextType(t *testing.T) {
	if !isContextType(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		t.Error("isContextType should return true for context.Context")
	}
	if isContextType(reflect.TypeOf("")) {
		t.Error("isContextType should return false for string")
	}
}

func TestIsErrorType(t *testing.T) {
	if !isErrorType(reflect.TypeOf((*error)(nil)).Elem()) {
		t.Error("isErrorType should return true for error")
	}
	if isErrorType(reflect.TypeOf("")) {
		t.Error("isErrorType should return false for string")
	}
}

func TestDecodeJSON(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name" path:"id"`
	}

	type args struct {
		req *http.Request
	}
	type want struct {
		v     TestStruct
		error bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid_json",
			args: args{
				req: func() *http.Request {
					req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"test"}`))
					req.Header.Set("Content-Type", "application/json")
					return req
				}(),
			},
			want: want{
				v:     TestStruct{Name: "test"},
				error: false,
			},
		},
		{
			name: "invalid_json",
			args: args{
				req: func() *http.Request {
					req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":}`))
					req.Header.Set("Content-Type", "application/json")
					return req
				}(),
			},
			want: want{
				v:     TestStruct{},
				error: true,
			},
		},
		{
			name: "with_path_param",
			args: args{
				req: func() *http.Request {
					req := httptest.NewRequest("POST", "/test/123", strings.NewReader(`{}`))
					req.Header.Set("Content-Type", "application/json")
					rctx := chi.NewRouteContext()
					rctx.URLParams.Add("id", "123")
					req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
					return req
				}(),
			},
			want: want{
				v:     TestStruct{Name: "123"},
				error: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v TestStruct
			err := decodeJSON(tt.args.req, &v)

			if tt.want.error {
				if err == nil {
					t.Error("Expected error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if v != tt.want.v {
					t.Errorf("Expected %v, got %v", tt.want.v, v)
				}
			}
		})
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	v := map[string]string{"key": "value"}
	writeJSON(w, nil, http.StatusOK, v)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Error("Content-Type not set correctly")
	}
	var resp map[string]string
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp["key"] != "value" {
		t.Error("JSON not written correctly")
	}
}

func TestDefaultConfigFor(t *testing.T) {
	tests := []struct {
		method string
		status int
	}{
		{"POST", http.StatusCreated},
		{"GET", http.StatusOK},
		{"PUT", http.StatusOK},
		{"DELETE", http.StatusOK},
	}

	for _, tt := range tests {
		cfg := defaultConfigFor(tt.method)
		if cfg.SuccessStatus != tt.status {
			t.Errorf("For %s, expected %d, got %d", tt.method, tt.status, cfg.SuccessStatus)
		}
	}
}
