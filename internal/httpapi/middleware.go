package httpapi

import (
	"net/http"
	"time"
)

func timeout(next http.Handler, d time.Duration) http.Handler {
	return http.TimeoutHandler(next, d, "request timeout")
}
func allowMethods(next http.Handler, methods ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, method := range methods {
			if r.Method == method {
				next.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("Allow", methods[0])
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}
