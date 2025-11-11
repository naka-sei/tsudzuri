package error

import (
	"encoding/json"
	"net/http"

	perrcode "github.com/naka-sei/tsudzuri/presentation/http/errcode"
	presp "github.com/naka-sei/tsudzuri/presentation/http/response"
)

const jsonContentType = "application/json; charset=utf-8"

// WriteError writes a JSON error response using the shared conversion logic.
func WriteError(w http.ResponseWriter, err error) {
	status, payload := convertError(err)

	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// convertError translates a domain/application error into an HTTP status code and payload.
func convertError(err error) (int, presp.ErrResponse) {
	if err == nil {
		return http.StatusInternalServerError, presp.ErrResponse{Message: "不明なエラーが発生しました。再度お試しください。"}
	}

	status := perrcode.GetStatusCode(err)
	reason := perrcode.GetErrorReason(err)
	if reason == nil {
		return status, presp.ErrResponse{Message: "不明なエラーが発生しました。再度お試しください。"}
	}

	return status, presp.ErrResponse{Message: reason.Message}
}
