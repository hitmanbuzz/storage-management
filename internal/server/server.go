package server

import (
	"log"

	"github.com/gin-gonic/gin"
)

type server struct {
	ip_addr string
	engine  *gin.Engine
}

func NewServer(ip_addr string) *server {
	return &server{
		ip_addr: ip_addr,
		engine:  gin.Default(),
	}
}

func (s *server) Routes() {
	// route := newRoute(s.engine)
}

func (s *server) Run() {
	err := s.engine.Run(s.ip_addr)
	if err != nil {
		log.Fatal(err)
	}
}
