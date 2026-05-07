package middleware

import (
	"log"
	"net/http"
)

func LogRequest(next http.Handler) http.Handler {
	return middleware(func(_ http.ResponseWriter, request *http.Request) {
		log.Printf("[%s] %s", request.Method, request.URL.Path)
	}, next)
}

type Header = [2]string

func AddHeaders(next http.Handler, headers ...Header) http.Handler {
	return middleware(func(response http.ResponseWriter, _ *http.Request) {
		for _, header := range headers {
			response.Header().Add(header[0], header[1])
		}
	}, next)
}

func NoCache(next http.Handler) http.Handler {
	return middleware(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Add("Cache-Control", "no-cache")
	}, next)
}

func middleware(fn http.HandlerFunc, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		fn(response, request)
		next.ServeHTTP(response, request)
	})
}
