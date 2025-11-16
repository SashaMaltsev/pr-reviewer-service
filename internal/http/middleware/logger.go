package custom_middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

/*

HTTP logging middleware for request tracking.
Logs method, path, status code and response time for each request.

*/

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(ww, r)

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", ww.statusCode).
			Dur("duration_ms", time.Since(start)).
			Msg("HTTP request")
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// overriding http.ResponseWriter method
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
