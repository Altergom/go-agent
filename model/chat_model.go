package model

import (
	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/openai"
)

type ChatModel struct {
	Ark    ark.ChatModel
	OpenAI openai.ChatModel
}
