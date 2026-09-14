package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	if w.statusCode == 0 {
		w.statusCode = code
	}

	w.ResponseWriter.WriteHeader(code)
}

func Logging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		writer := &statusResponseWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(writer, r)

		status := writer.statusCode
		if status == 0 {
			status = http.StatusOK
		}

		requestID := GetRequestID(r.Context())

		duration := time.Since(start)

		logger.Info(
			"request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"duration", duration,
			"request_id", requestID)
	})
}
