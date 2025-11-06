package server

import (
	"context"
	"fmt"
	"mango/internal/controller"
	"net/http"
	"time"

	"mango/config"

	"github.com/gin-gonic/gin"
)

// Server HTTP 服务器
type Server struct {
	config     *config.Config
	engine     *gin.Engine
	httpServer *http.Server
	handlers   []controller.Handler
}

// NewServer 创建 HTTP 服务器
func NewServer(config *config.Config, handlers ...controller.Handler) *Server {
	engine := gin.Default()

	// 注册全局中间件
	engine.Use(gin.Recovery())

	return &Server{
		config:   config,
		engine:   engine,
		handlers: handlers,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	// 注册路由
	router := s.engine.Group("/api")
	for _, h := range s.handlers {
		h.Register(router)
	}

	// 创建 HTTP 服务器
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.engine,
	}

	return s.httpServer.ListenAndServe()
}

// Stop 停止服务器
func (s *Server) Stop() error {
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
