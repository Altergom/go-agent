package main

import (
	"go-agent/config"
	"log"
)

func main() {
	// 初始化config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("警告: 未找到 .env 文件")
	}
	config.Cfg = cfg

	// 初始化模型
}
