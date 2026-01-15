package model

import (
	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/model"
)

type Model struct {
	ChatModel      ChatModel
	EmbeddingModel EmbeddingModel
}

var Md *Model

func LoadChatModel(modelType string) (*Model, error) {
	switch modelType {
	case "ark":
		Md.ChatModel.Ark = ark.ChatModel{}
	}
}
