package server

import (
	"github.com/gorilla/mux"
)

//Handler is a contract to every handler that want to be included in the apps
type Handler interface {
	RegisterRoutes(router *mux.Router)
}
