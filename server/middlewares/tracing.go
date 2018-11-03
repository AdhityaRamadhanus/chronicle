package middlewares

import (
	"net/http"

	"github.com/google/uuid"
)

//TraceRequest add X-Request-ID header to request
func TraceRequest(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		req.Header.Set("X-Request-ID", uuid.New().String())
		res.Header().Set("X-Request-ID", uuid.New().String())

		nextHandler.ServeHTTP(res, req)
	})
}
