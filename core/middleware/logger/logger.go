package logger

import (
	"log"
	"net/http"
)

func Logger(prefix string) func(
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) {
			log.Printf(
				"%s Method: %s, path: %s",
				prefix,
				request.Method,
				request.URL.Path,
			)
			next(writer, request)
		}
	}
}
