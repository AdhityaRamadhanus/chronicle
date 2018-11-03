package middlewares

import (
	"net/http"
	"runtime/debug"

	"github.com/AdhityaRamadhanus/chronicle/server/render"
	log "github.com/sirupsen/logrus"
)

//PanicHandler will return statusCOde 500 and log panic error
func PanicHandler(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.WithError(err.(error)).Errorf("Panic error %s", debug.Stack())
				render.JSON(res, http.StatusInternalServerError, map[string]interface{}{
					"status": http.StatusInternalServerError,
					"error": map[string]interface{}{
						"code":    "ErrInternalServer",
						"message": "Something is Wrong",
					},
				})
			}
		}()
		nextHandler.ServeHTTP(res, req)
	})
}
