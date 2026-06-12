package server

import (
	"log"
	"log/slog"

	"github.com/gin-gonic/gin"
)

type server struct {
	logger  *slog.Logger
	ip_addr string
	engine  *gin.Engine
}

func NewServer(ip_addr string, logger *slog.Logger) *server {
	return &server{
		logger:  logger,
		ip_addr: ip_addr,
		engine:  gin.Default(),
	}
}

func (s *server) Routes() {
	routes := newRoutes(s.engine, s.logger)
	s.engine.POST("/upload", routes.Upload)
}

func (s *server) Run() {
	err := s.engine.Run(s.ip_addr)
	if err != nil {
		log.Fatal(err)
	}

	s.logger.Debug("server running", "ip", s.ip_addr)
}
