package custom_middleware

import (
	"net/http"

	"github.com/rs/zerolog/log"
)


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
                w.Write([]byte(`
                    {
                        "error":
                            {
                                "code":"INTERNAL_ERROR",
                                "message":"internal server error"
                            }
                    }
                    `))
            }
        }()

        next.ServeHTTP(w, r)
    })
}