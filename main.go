package main

import (
	"go-agent/config"
	"log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("警告: 未找到 .env 文件")
	}
	config.Cfg = cfg
}
