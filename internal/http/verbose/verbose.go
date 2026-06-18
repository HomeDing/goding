package verbose

// verbose logging middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// HttpLogging prints method, path, and request duration
func HttpLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("http",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("content-type", w.Header().Get("Content-Type")),
			slog.Duration("duration", time.Since(start)))
	})
}
