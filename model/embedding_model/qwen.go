package embedding_model

import (
	"context"
	"fmt"
	"go-agent/config"
	"net/http"
	"time"

	"github.com/cloudwego/eino-ext/libs/acl/openai"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/embedding"
)

var (
	defaultBaseUrl = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	defaultDim     = 2048
)

type EmbeddingConfig struct {
	APIKey     string        `json:"api_key"`
	BaseURL    string        `json:"base_url"`
	Timeout    time.Duration `json:"timeout"`
	HTTPClient *http.Client  `json:"http_client"`
	Model      string        `json:"model"`
	// TODO 因为vllm不支持千问3的MRL，自定义维度会导致使用千问3报错。暂时屏蔽等待vllm支持
	//Dimensions *int          `json:"dimensions,omitempty"`
}

type Embedder struct {
	cli *openai.EmbeddingClient
}

func NewEmbedder(ctx context.Context, config *EmbeddingConfig) (*Embedder, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	var httpClient *http.Client
	if config.HTTPClient != nil {
		httpClient = config.HTTPClient
	} else {
		httpClient = &http.Client{Timeout: config.Timeout}
	}

	// 千问配置映射到OpenAI
	cfg := &openai.EmbeddingConfig{
		APIKey:     config.APIKey,
		BaseURL:    config.BaseURL,
		HTTPClient: httpClient,
		Model:      config.Model,
		//Dimensions: &defaultDim,
	}

	cli, err := openai.NewEmbeddingClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &Embedder{cli: cli}, nil
}

func (e *Embedder) EmbedStrings(ctx context.Context, text []string, opts ...embedding.Option) ([][]float64, error) {
	ctx = callbacks.EnsureRunInfo(ctx, e.GetType(), components.ComponentOfEmbedding)

	return e.cli.EmbedStrings(ctx, text, opts...)
}

const typ = "Qwen"

func (e *Embedder) GetType() string {
	return typ
}

func (e *Embedder) IsCallbacksEnabled() bool {
	return true
}

func initQwen() {
	registerEmbeddingModel("qwen", func(ctx context.Context) (embedding.Embedder, error) {
		if config.Cfg.QwenConf.BaseUrl != "" {
			defaultBaseUrl = config.Cfg.QwenConf.BaseUrl
		}
		emb, err := NewEmbedder(ctx, &EmbeddingConfig{
			APIKey:  config.Cfg.QwenConf.QwenKey,
			Model:   config.Cfg.QwenConf.QwenEmbedding,
			BaseURL: defaultBaseUrl,
		})
		if err != nil {
			return nil, err
		}

		return emb, nil
	})
}
