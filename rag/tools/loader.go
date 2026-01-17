package tools

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino/components/document"
)

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
