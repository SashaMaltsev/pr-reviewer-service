package custom_middleware

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

// Recovery is a middleware that recovers from panics, logs the error,
// and returns a 500 Internal Server Error response to the client.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			err := recover()

			if err != nil {
				log.Error().
					Interface("error", err).
					Str("path", r.URL.Path).
					Msg("Panic recovered")

				w.WriteHeader(http.StatusInternalServerError)
				_, writeErr := w.Write([]byte(`{
					"error": {
						"code": "INTERNAL_ERROR",
						"message": "internal server error"
					}
				}`))

				if writeErr != nil {
					log.Error().
						Err(writeErr).
						Str("path", r.URL.Path).
						Msg("Failed to write error response")
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
