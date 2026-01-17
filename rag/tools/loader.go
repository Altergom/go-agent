package tools

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino/components/document"
)

// Loader 文档加载器，提供Load方法将文档转为[]*schema.Document数据结构
var Loader document.Loader

func NewLoader(ctx context.Context) (document.Loader, error) {
	loader, err := file.NewFileLoader(ctx, &file.FileLoaderConfig{
		UseNameAsID: true,
	})
	if err != nil {
		return nil, err
	}

	return loader, nil
}
