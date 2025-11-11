package error

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	presp "github.com/naka-sei/tsudzuri/presentation/http/response"
)

func TestConvertError(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		err          error
		wantStatus   int
		wantResponse presp.ErrResponse
	}{
		"known user error": {
			err:        duser.ErrUserNotFound,
			wantStatus: http.StatusUnauthorized,
			wantResponse: presp.ErrResponse{
				Message: "認証が必要です。ログインしてください。",
			},
		},
		"unknown error": {
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantResponse: presp.ErrResponse{
				Message: "不明なエラーが発生しました。再度お試しください。",
			},
		},
		"nil error": {
			err:        nil,
			wantStatus: http.StatusInternalServerError,
			wantResponse: presp.ErrResponse{
				Message: "不明なエラーが発生しました。再度お試しください。",
			},
		},
	}

	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotStatus, gotResponse := convertError(tt.err)
			if gotStatus != tt.wantStatus {
				t.Fatalf("convertError() status = %d, want %d", gotStatus, tt.wantStatus)
			}
			if gotResponse != tt.wantResponse {
				t.Fatalf("convertError() response = %#v, want %#v", gotResponse, tt.wantResponse)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	WriteError(rec, duser.ErrUserNotFound)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("WriteError() status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	if got := rec.Header().Get("Content-Type"); got != jsonContentType {
		t.Fatalf("WriteError() content-type = %q, want %q", got, jsonContentType)
	}

	var payload presp.ErrResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	want := presp.ErrResponse{Message: "認証が必要です。ログインしてください。"}
	if payload != want {
		t.Fatalf("WriteError() payload = %#v, want %#v", payload, want)
	}
}
