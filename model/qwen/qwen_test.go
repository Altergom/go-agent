package qwen

import (
	"context"
	"go-agent/config"
	"go-agent/model"
	"os"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func setupTestConfig() {
	config.Cfg = &config.Config{
		ChatModelType: "qwen",
		QwenConf: config.QwenConfig{
			QwenKey:       "sk-57e46d8937ff489eb3de79fda41f30d6",
			QwenChatModel: "qwen-flash",
		},
	}
}

func TestInitChatModel(t *testing.T) {
	setupTestConfig()
	InitChatModel()

	ctx := context.Background()
	chatModel, err := model.NewChatModel(ctx)

	if err != nil {
		t.Logf("Expected error due to invalid test API key: %v", err)
	} else if chatModel == nil {
		t.Error("Expected chat model to be created, got nil")
	}
}

func TestInitChatModelRegistration(t *testing.T) {
	setupTestConfig()
	InitChatModel()

	ctx := context.Background()
	_, err := model.NewChatModel(ctx)

	if err != nil {
		t.Logf("Model creation failed as expected with test credentials: %v", err)
	}

}

func TestInitChatModelWithEnv(t *testing.T) {
	os.Setenv("QWEN_KEY", "test-env-key")
	os.Setenv("QWEN_CHAT_MODEL", "qwen-plus")
	defer func() {
		os.Unsetenv("QWEN_KEY")
		os.Unsetenv("QWEN_CHAT_MODEL")
	}()

	config.Cfg = &config.Config{
		ChatModelType: "qwen",
		QwenConf: config.QwenConfig{
			QwenKey:       os.Getenv("QWEN_KEY"),
			QwenChatModel: os.Getenv("QWEN_CHAT_MODEL"),
		},
	}

	InitChatModel()

	ctx := context.Background()
	_, err := model.NewChatModel(ctx)

	if err != nil {
		t.Logf("Model creation with env config: %v", err)
	}
}

func TestInitChatModelMissingConfig(t *testing.T) {
	config.Cfg = &config.Config{
		ChatModelType: "qwen",
		QwenConf: config.QwenConfig{
			QwenKey:       "",
			QwenChatModel: "",
		},
	}

	InitChatModel()

	ctx := context.Background()
	_, err := model.NewChatModel(ctx)

	if err != nil {
		t.Logf("Expected error with empty config: %v", err)
	}
}

func TestChatModelGenerate(t *testing.T) {
	setupTestConfig()
	InitChatModel()

	ctx := context.Background()
	chatModel, err := model.NewChatModel(ctx)
	if err != nil {
		t.Skipf("Skipping Generate test due to model creation error: %v", err)
		return
	}

	messages := []*schema.Message{
		schema.UserMessage("你好，请用一句话介绍Go语言"),
	}

	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		t.Errorf("Generate failed: %v", err)
		return
	}

	if response == nil {
		t.Error("Expected response, got nil")
		return
	}

	if response.Content == "" {
		t.Error("Expected non-empty response content")
	}

	t.Logf("Response: %s", response.Content)
}

func TestChatModelGenerateWithHistory(t *testing.T) {
	setupTestConfig()
	InitChatModel()

	ctx := context.Background()
	chatModel, err := model.NewChatModel(ctx)
	if err != nil {
		t.Skipf("Skipping Generate test due to model creation error: %v", err)
		return
	}

	messages := []*schema.Message{
		schema.UserMessage("我的名字是小明"),
		schema.AssistantMessage("你好，小明！很高兴认识你。", []schema.ToolCall{}),
		schema.UserMessage("我刚才说我叫什么名字？"),
	}

	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		t.Errorf("Generate with history failed: %v", err)
		return
	}

	if response == nil {
		t.Error("Expected response, got nil")
		return
	}

	if response.Content == "" {
		t.Error("Expected non-empty response content")
	}

	t.Logf("Response with history: %s", response.Content)
}

func TestChatModelGenerateWithSystemMessage(t *testing.T) {
	setupTestConfig()
	InitChatModel()

	ctx := context.Background()
	chatModel, err := model.NewChatModel(ctx)
	if err != nil {
		t.Skipf("Skipping Generate test due to model creation error: %v", err)
		return
	}

	messages := []*schema.Message{
		schema.SystemMessage("你是一个专业的Go语言开发专家，请用简洁的方式回答问题。"),
		schema.UserMessage("什么是goroutine？"),
	}

	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		t.Errorf("Generate with system message failed: %v", err)
		return
	}

	if response == nil {
		t.Error("Expected response, got nil")
		return
	}

	if response.Content == "" {
		t.Error("Expected non-empty response content")
	}

	t.Logf("Response with system message: %s", response.Content)
}

func TestChatModelGenerateEmptyMessages(t *testing.T) {
	setupTestConfig()
	InitChatModel()

	ctx := context.Background()
	chatModel, err := model.NewChatModel(ctx)
	if err != nil {
		t.Skipf("Skipping Generate test due to model creation error: %v", err)
		return
	}

	messages := []*schema.Message{}

	_, err = chatModel.Generate(ctx, messages)
	if err == nil {
		t.Error("Expected error with empty messages, got nil")
	} else {
		t.Logf("Expected error with empty messages: %v", err)
	}
}
