package app

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"mango/config"
	"mango/internal/server"

	"github.com/sirupsen/logrus"
)

// App 应用程序
type App struct {
	config *config.Config
	server *server.Server
}

// NewApp 创建应用程序
func NewApp(config *config.Config, server *server.Server) *App {
	return &App{
		config: config,
		server: server,
	}
}

// Run 运行应用程序
func (a *App) Run() error {
	// 配置日志
	if a.config.Log.Level != "" {
		level, err := logrus.ParseLevel(a.config.Log.Level)
		if err == nil {
			logrus.SetLevel(level)
		}
	}

	// 创建退出通道
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 启动 HTTP 服务器
	go func() {
		if err := a.server.Start(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待退出信号
	<-quit
	logrus.Info("Shutting down server...")

	// 关闭服务器
	if err := a.server.Stop(); err != nil {
		logrus.Errorf("Server forced to shutdown: %v", err)
	}

	logrus.Info("Server exited")
	return nil
}
