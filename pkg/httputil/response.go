package httputil

import (
	"context"
	"net/http"
	"reflect"
	"time"

	"github.com/bytedance/sonic"
)

func init() {
	// Pre-compile JIT encoder/decoder for response types at startup.
	// Eliminates first-request JIT penalty.
	sonic.Pretouch(reflect.TypeOf(Response{}))
}

type Response struct {
	Status     string      `json:"status"`
	Data       any         `json:"data,omitempty"`
	Error      *ErrorBody  `json:"error,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Meta       Meta        `json:"meta"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

type Pagination struct {
	Cursor  string `json:"cursor,omitempty"`
	Limit   int    `json:"limit"`
	HasNext bool   `json:"has_next"`
}

type Meta struct {
	RequestID string `json:"request_id"`
	Timestamp int64  `json:"timestamp"`
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	buf, err := sonic.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(buf)
}

func Success(w http.ResponseWriter, r *http.Request, status int, data any) {
	writeJSON(w, status, Response{
		Status: "success",
		Data:   data,
		Meta:   buildMeta(r),
	})
}

func SuccessList(w http.ResponseWriter, r *http.Request, data any, pagination Pagination) {
	writeJSON(w, http.StatusOK, Response{
		Status:     "success",
		Data:       data,
		Pagination: &pagination,
		Meta:       buildMeta(r),
	})
}

func Error(w http.ResponseWriter, r *http.Request, status int, code string, message string) {
	writeJSON(w, status, Response{
		Status: "error",
		Error:  &ErrorBody{Code: code, Message: message},
		Meta:   buildMeta(r),
	})
}

func buildMeta(r *http.Request) Meta {
	reqID := GetRequestID(r.Context())
	if reqID == "" {
		reqID = r.Header.Get("X-Request-ID")
	}
	return Meta{
		RequestID: reqID,
		Timestamp: time.Now().UnixMilli(),
	}
}

type requestIDKey struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}
