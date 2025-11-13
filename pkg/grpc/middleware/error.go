package middleware

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/naka-sei/tsudzuri/pkg/log"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorBody represents the external HTTP error response converted from a gRPC status.
type ErrorBody struct {
	ErrorCode     string `json:"errorCode"`
	Message       string `json:"message"`
	ClientMessage string `json:"clientMessage"`
}

// NewErrorHandler creates a grpc-gateway error handler that maps gRPC status errors to HTTP error responses.
func NewErrorHandler() runtime.ErrorHandlerFunc {
	return func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
		l := log.LoggerFromContext(ctx)

		st, ok := status.FromError(err)
		if !ok {
			runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
			return
		}

		var errInfo *errdetails.ErrorInfo
		for _, detail := range st.Details() {
			if t, ok := detail.(*errdetails.ErrorInfo); ok {
				errInfo = t
				break
			}
		}

		if errInfo == nil {
			runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
			return
		}

		body := ErrorBody{
			ErrorCode:     errInfo.Reason,
			Message:       st.Message(),
			ClientMessage: errInfo.Metadata["client_message"],
		}

		payload, marshalErr := marshaler.Marshal(&body)
		if marshalErr != nil {
			runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
			return
		}

		if st.Code() == codes.Unauthenticated {
			w.Header().Set("WWW-Authenticate", st.Message())
		}
		w.Header().Set("Content-Type", marshaler.ContentType(body))
		statusCode := runtime.HTTPStatusFromCode(st.Code())
		w.WriteHeader(statusCode)
		if _, err = w.Write(payload); err != nil {
			l.Sugar().Errorf("failed to write error response: %v", err)
			runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
			return
		}
	}
}
