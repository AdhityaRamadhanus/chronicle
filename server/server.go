package server

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/AdhityaRamadhanus/chronicle/server/middlewares"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

//Server hold mux Router and information of host port and address of our app
type Server struct {
	Router *mux.Router
	Port   string
}

//NewServer create Server from Handler
func NewServer(Handlers []Handler) *Server {
	router := mux.NewRouter().
		StrictSlash(true).
		PathPrefix("/api").
		Subrouter()

	for _, handler := range Handlers {
		handler.RegisterRoutes(router)
	}

	return &Server{
		Router: router,
		Port:   os.Getenv("PORT"),
	}
}

//CreateHttpServer will return http.Server for flexible use like testing
func (s *Server) CreateHttpServer() *http.Server {
	srv := &http.Server{
		Handler: middlewares.PanicHandler(
			middlewares.Gzip(
				middlewares.TraceRequest(
					cors.Default().Handler(
						middlewares.LogRequest(s.Router),
					),
				),
			),
		),
		Addr:         fmt.Sprintf(":%d", os.Getenv("PORT")),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	return srv
}
