//go:build wireinject
// +build wireinject

package main

import (
	"fmt"
	"mango/internal/controller"

	"mango/config"
	"mango/internal/app"
	"mango/internal/repository"
	"mango/internal/server"
	"mango/internal/service"

	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitializeApp 初始化应用程序
func InitializeApp(configPath string) (*app.App, error) {
	wire.Build(
		// 配置
		config.NewConfig,

		// 数据库
		provideDB,

		// 仓库
		repository.NewUserRepository,
		repository.NewTextRiskLogRepository,
		//wire.Bind(new(repository.UserRepository), new(*repository.UserRepositoryS)),

		// 服务
		service.NewUserService,
		//wire.Bind(new(service.UserService), new(*service.UserServiceS)),
		service.NewTextRiskLogService,

		// 处理器
		controller.NewUserHandler,
		provideHandlers,
		controller.NewVolcHandler,
		controller.NewVoiceHandler,
		controller.NewZhiPuHandler,
		controller.NewAlgorithmHandler,

		// 服务器
		server.NewServer,

		// 应用程序
		app.NewApp,
	)
	return nil, nil
}

// provideDB 提供数据库连接
func provideDB(config *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Database.Username,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.DBName,
	)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

func provideHandlers(userHandler *controller.UserHandler, volcHandler *controller.VolcHandler, voiceHandler *controller.VoiceHandler, zhipuHandler *controller.ZhiPuHandler, algorithmHandler *controller.AlgorithmHandler) []controller.Handler {
	return []controller.Handler{userHandler, volcHandler, voiceHandler, zhipuHandler, algorithmHandler}
}
