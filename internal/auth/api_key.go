package auth

import (
	"crypto/subtle"
	"net/http"
)

const DefaultHeader = "X-API-Key"

func RequireAPIKey(next http.Handler, apiKey, header string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := apiKey
		if expected == "" {
			http.Error(w, "missing api key configuration", http.StatusUnauthorized)
			return
		}
		name := header
		if name == "" {
			name = DefaultHeader
		}
		provided := r.Header.Get(name)
		if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			http.Error(w, "invalid api key", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireAPIKeyHandler(apiKey, header string, handler http.HandlerFunc) http.Handler {
	return RequireAPIKey(handler, apiKey, header)
}
