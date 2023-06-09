package middleware

import (
	"fmt"
	"net/http"
)

func Panic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("\n\npanic middleware", r.URL.Path)
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("recovered with err:", err)
				http.Error(w, "Internal server error", 500)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
