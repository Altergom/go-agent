package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// 模型类型配置
	ChatModelType      string
	EmbeddingModelType string

	ArkConf    ArkConfig
	OpenAIConf OpenAIConfig
}

type ArkConfig struct {
	ArkKey            string
	ArkEmbeddingModel string
	ArkChatModel      string
}

type OpenAIConfig struct {
	OpenAIKey       string
	OpenAIChatModel string
}

var Cfg *Config

func LoadConfig() (*Config, error) {
	// 加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	config := &Config{
		ChatModelType:      getEnv("CHAT_MODEL_TYPE", "ark"),
		EmbeddingModelType: getEnv("EMBEDDING_MODEL_TYPE", "ark"),

		ArkConf: ArkConfig{
			ArkKey:            getEnv("ARK_KEY", ""),
			ArkEmbeddingModel: getEnv("ARK_EMBEDDING_MODEL", ""),
			ArkChatModel:      getEnv("ARK_CHAT_MODEL", ""),
		},
		OpenAIConf: OpenAIConfig{
			OpenAIKey:       getEnv("OPENAI_KEY", ""),
			OpenAIChatModel: getEnv("OPENAI_CHAT_MODEL", "gpt-4"),
		},
	}

	return config, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
