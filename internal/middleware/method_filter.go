package middleware

import (
	"net/http"
	"slices"
)

func MethodFilter(methods []string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if slices.Contains(methods, r.Method) {
			next(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
