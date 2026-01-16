package model

import (
	"context"
	"fmt"
	"go-agent/config"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type ChatModel struct {
	ModelType string
	Ark       *ark.ChatModel
	OpenAI    *openai.ChatModel
}

var CM *ChatModel

// NewChatModel 根据配置创建 ChatModel 实例
func NewChatModel(ctx context.Context) (*ChatModel, error) {
	cm := &ChatModel{
		ModelType: config.Cfg.ChatModelType,
	}

	// 根据配置初始化对应的模型
	switch config.Cfg.ChatModelType {
	case "ark":
		arkModel, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
			APIKey: config.Cfg.ArkConf.ArkKey,
			Model:  config.Cfg.ArkConf.ArkChatModel,
		})
		if err != nil {
			return nil, fmt.Errorf("初始化 Ark ChatModel 失败: %v", err)
		}
		cm.Ark = arkModel

	case "openai":
		openaiModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey: config.Cfg.OpenAIConf.OpenAIKey,
			Model:  config.Cfg.OpenAIConf.OpenAIChatModel,
		})
		if err != nil {
			return nil, fmt.Errorf("初始化 OpenAI ChatModel 失败: %v", err)
		}
		cm.OpenAI = openaiModel

	default:
		return nil, fmt.Errorf("不支持的 ChatModel 类型: %s", config.Cfg.ChatModelType)
	}

	return cm, nil
}

// Generate 常规输出方法
func (cm ChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	switch cm.ModelType {
	case "ark":
		// 检查 arkModel 是否已初始化（接口类型的零值检查）
		var zeroArk *ark.ChatModel
		if cm.Ark == zeroArk {
			return nil, fmt.Errorf("Ark ChatModel 未初始化")
		}
		return cm.Ark.Generate(ctx, input, opts...)

	case "openai":
		// 检查 openaiModel 是否已初始化（接口类型的零值检查）
		var zeroOpenAI *openai.ChatModel
		if cm.OpenAI == zeroOpenAI {
			return nil, fmt.Errorf("OpenAI ChatModel 未初始化")
		}
		return cm.OpenAI.Generate(ctx, input, opts...)

	default:
		return nil, fmt.Errorf("不支持的 ChatModel 类型: %s", cm.ModelType)
	}
}

// Stream 流式输出方法
func (cm ChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	switch cm.ModelType {
	case "ark":
		// 初始化检查
		var zeroArk *ark.ChatModel
		if cm.Ark == zeroArk {
			return nil, fmt.Errorf("Ark ChatModel 未初始化")
		}
		// Ark 模型支持流式输出，可以添加缓存选项
		arkOpts := make([]model.Option, 0, len(opts)+1)
		arkOpts = append(arkOpts, opts...)
		return cm.Ark.Stream(ctx, input, arkOpts...)

	case "openai":
		// 初始化检查
		var zeroOpenAI *openai.ChatModel
		if cm.OpenAI == zeroOpenAI {
			return nil, fmt.Errorf("OpenAI ChatModel 未初始化")
		}
		return cm.OpenAI.Stream(ctx, input, opts...)

	default:
		return nil, fmt.Errorf("不支持的 ChatModel 类型: %s", cm.ModelType)
	}
}
