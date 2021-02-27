package middleware

import (
	"log"
	"net/http"
	"time"
)

func ElapsedTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeStart := time.Now()
		// next handler
		next.ServeHTTP(w, r)
		timeElapsed := time.Since(timeStart)
		log.Println("handle", r.URL.Path, timeElapsed)
	})
}
