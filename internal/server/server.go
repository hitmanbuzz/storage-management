package server

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	logger     *slog.Logger
	ip_addr    string
	engine     *gin.Engine
	httpServer *http.Server
}

func NewServer(ip_addr string, logger *slog.Logger) *Server {
	engine := gin.Default()
	server := &http.Server{
		Addr:    ip_addr,
		Handler: engine,
	}

	return &Server{
		logger:     logger,
		ip_addr:    ip_addr,
		engine:     engine,
		httpServer: server,
	}
}

func (s *Server) Routes() {
	routes := newRoutes(s.engine, s.logger)
	s.engine.POST("/upload", routes.Upload)
}

func (s *Server) Run() error {
	s.logger.Debug("server running", "ip", s.ip_addr)
	s.Routes()
	err := s.httpServer.ListenAndServe()
	return err
}

func (s *Server) GetHttpServer() *http.Server {
	return s.httpServer
}
