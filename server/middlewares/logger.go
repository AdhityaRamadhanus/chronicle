package middlewares

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func LogRequest(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		start := time.Now()
		nextHandler.ServeHTTP(res, req)
		log.WithFields(log.Fields{
			"method":    req.Method,
			"resp time": time.Since(start),
		}).Info("PATH " + req.URL.Path)
	})
}
