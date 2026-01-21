package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// 模型类型配置
	ChatModelType      string
	EmbeddingModelType string
	VectorDBType       string

	ArkConf    ArkConfig
	OpenAIConf OpenAIConfig
	QwenConf   QwenConfig

	MilvusConf MilvusConfig
	ESConf     ESConfig
}

type ArkConfig struct {
	ArkKey            string
	ArkEmbeddingModel string
	ArkChatModel      string
}

type OpenAIConfig struct {
	OpenAIKey       string
	OpenAIChatModel string
	OpenAIEmbedding string
}

type QwenConfig struct {
	BaseUrl       string
	QwenKey       string
	QwenChatModel string
	QwenEmbedding string
}

type MilvusConfig struct {
	MilvusAddr          string
	MilvusUserName      string
	MilvusPassword      string
	SimilarityThreshold string
	CollectionName      string
	TopK                string
}

type ESConfig struct {
	Addresses []string
	Username  string
	Password  string
	CloudID   string
	APIKey    string
	Index     string
}

var Cfg *Config

func LoadConfig() (*Config, error) {
	// 加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}
	rawAddr := getEnv("ES_ADDRESS", "http://localhost:9200")

	// 按逗号分割成 []string
	esAddresses := strings.Split(rawAddr, ",")

	config := &Config{
		ChatModelType:      getEnv("CHAT_MODEL_TYPE", "ark"),
		EmbeddingModelType: getEnv("EMBEDDING_MODEL_TYPE", "ark"),
		VectorDBType:       getEnv("VECTOR_DB_TYPE", "milvus"),

		ArkConf: ArkConfig{
			ArkKey:            getEnv("ARK_KEY", ""),
			ArkEmbeddingModel: getEnv("ARK_EMBEDDING_MODEL", "doubao-embedding-text-240715"),
			ArkChatModel:      getEnv("ARK_CHAT_MODEL", "doubao-seed-1-8-251228"),
		},
		OpenAIConf: OpenAIConfig{
			OpenAIKey:       getEnv("OPENAI_KEY", ""),
			OpenAIChatModel: getEnv("OPENAI_CHAT_MODEL", "gpt-4"),
			OpenAIEmbedding: getEnv("OPENAI_EMBEDDING_MODEL", ""),
		},
		QwenConf: QwenConfig{
			BaseUrl:       getEnv("QWEN_BASE_URL", ""),
			QwenKey:       getEnv("QWEN_KEY", ""),
			QwenEmbedding: getEnv("QWEN_EMBEDDING_MODEL", ""),
			QwenChatModel: getEnv("QWEN_CHAT_MODEL", ""),
		},
		MilvusConf: MilvusConfig{
			MilvusAddr:          getEnv("MILVUS_ADDR", "localhost:27017"),
			MilvusUserName:      getEnv("MILVUS_USERNAME", ""),
			MilvusPassword:      getEnv("MILVUS_PASSWORD", ""),
			SimilarityThreshold: getEnv("MILVUS_SIMILARITY_THRESHOLD", "0.7"),
			CollectionName:      getEnv("MILVUS_COLLECTION_NAME", "GoAgent"),
			TopK:                getEnv("TOPK", "10"),
		},
		ESConf: ESConfig{
			Addresses: esAddresses,
			Username:  getEnv("ES_USERNAME", ""),
			Password:  getEnv("ES_PASSWORD", ""),
			Index:     getEnv("ES_INDEX", "go_agent_docs"),
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
