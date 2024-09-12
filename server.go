package todo

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func (server *Server) Run(port string, handler http.Handler) error {
	server.httpServer = &http.Server{
		Addr:           ":" + port,
		Handler:        handler,
		MaxHeaderBytes: 1048576,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
	}

	return server.httpServer.ListenAndServe()
}

func (server *Server) Shutdown(ctx context.Context) error {
	return server.httpServer.Shutdown(ctx)
}
