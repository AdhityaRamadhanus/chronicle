package server

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/adhityaramadhanus/chronicle/server/middlewares"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Server struct {
	Router *mux.Router
	Addr   string
	Host   string
	Port   string
}

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
		Host:   os.Getenv("CHRONICLE_HOST"),
		Port:   os.Getenv("CHRONICLE_PORT"),
		Addr:   fmt.Sprintf("%s:%s", os.Getenv("CHRONICLE_HOST"), os.Getenv("CHRONICLE_PORT")),
	}
}

func (s *Server) CreateHttpServer() *http.Server {
	srv := &http.Server{
		Handler: middlewares.PanicHandler(
			cors.Default().Handler(
				middlewares.LogRequest(s.Router),
			),
		),
		Addr:         s.Addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	return srv
}
