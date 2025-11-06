package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config/config.yaml", "Path to configuration file")
	flag.Parse()
	fmt.Println(configPath)
	// 初始化应用程序
	app, err := InitializeApp(*configPath)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// 运行应用程序
	if err := app.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
