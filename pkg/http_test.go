package pkg

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHttpHandler_ServeHTTP(t *testing.T) {
	// disable logs in logging middleware
	b := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(b, nil)))

	tests := []struct {
		name       string
		httpFunc   HttpHandler
		wantStatus int
	}{
		{
			name: "OK response status",
			httpFunc: HttpHandler(func(w http.ResponseWriter, r *http.Request) (any, error) {
				return "ok", nil
			}),
			wantStatus: http.StatusOK,
		},
		{
			name: "NoContent response when body is nil",
			httpFunc: HttpHandler(func(w http.ResponseWriter, r *http.Request) (any, error) {
				return nil, nil
			}),
			wantStatus: http.StatusNoContent,
		},
		{
			name: "ServiceUnavailableError",
			httpFunc: HttpHandler(func(w http.ResponseWriter, r *http.Request) (any, error) {
				return nil, NewServiceUnavailableError(errors.New("service unavailable"))
			}),
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name: "BadRequestError",
			httpFunc: HttpHandler(func(w http.ResponseWriter, r *http.Request) (any, error) {
				return nil, NewBadRequestError(InvalidParam{"param1", "error1"})
			}),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "InternalServerError",
			httpFunc: HttpHandler(func(w http.ResponseWriter, r *http.Request) (any, error) {
				return nil, NewInternalServerError(errors.New("internal error"))
			}),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "NotFoundError",
			httpFunc: HttpHandler(func(w http.ResponseWriter, r *http.Request) (any, error) {
				return nil, NewNotFoundError("not found")
			}),
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, _ := http.NewRequest(http.MethodGet, "/", nil)
			recorder := httptest.NewRecorder()

			tt.httpFunc.ServeHTTP(recorder, request)

			// Assert
			if status := recorder.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.wantStatus)
			}
		})
	}
}
