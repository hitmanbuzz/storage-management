package server

import (
	"log/slog"
	"net/http"
	"os"
	"storage-management/internal/auth"
	"storage-management/internal/database"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

type Server struct {
	logger     *slog.Logger
	ip_addr    string
	engine     *gin.Engine
	db         *database.DatabaseHandler
	httpServer *http.Server
}

func NewServer(ip_addr string, db *database.DatabaseHandler, logger *slog.Logger) *Server {
	engine := gin.Default()
	store := cookie.NewStore([]byte(os.Getenv("COOKIE_KEY")))
	engine.Use(sessions.Sessions("auth-session", store))

	server := &http.Server{
		Addr:    ip_addr,
		Handler: engine,
	}

	return &Server{
		logger:     logger,
		ip_addr:    ip_addr,
		engine:     engine,
		db:         db,
		httpServer: server,
	}
}

func (s *Server) Routes() {
	routes := newRoutes(s.engine, s.db, s.logger)
	s.engine.POST("/register", routes.Register)
	s.engine.POST("/login", routes.Login)

	auths := s.engine.Group("/")
	auths.Use(auth.AuthMiddleware)
	{
		auths.POST("/upload", routes.Upload)
	}
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
