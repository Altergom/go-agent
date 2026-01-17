package compose

import (
	"context"
	"fmt"
	"go-agent/rag/tools"
	"os"
	"strings"

	"github.com/cloudwego/eino/components/document/parser"
	"github.com/cloudwego/eino/schema"
)

func BuildParseNode(ctx context.Context, input []*schema.Document) ([]*schema.Document, error) {
	var parsedDocs []*schema.Document

	for _, doc := range input {
		uri, ok := doc.MetaData["uri"].(string)
		if !ok {
			uri, _ = doc.MetaData["source"].(string)
		}

		if uri != "" {
			file, err := os.Open(uri)
			if err != nil {
				return nil, fmt.Errorf("failed to open file: %w", err)
			}
			defer file.Close()

			parsed, err := tools.Parser.Parse(ctx, file,
				parser.WithURI(uri),
				parser.WithExtraMeta(doc.MetaData),
			)
			if err != nil {
				return nil, fmt.Errorf("failed to parse document: %w", err)
			}

			parsedDocs = append(parsedDocs, parsed...)
		} else {
			// 如果没有 URI，尝试从内容解析
			reader := strings.NewReader(doc.Content)
			parsed, err := tools.Parser.Parse(ctx, reader,
				parser.WithExtraMeta(doc.MetaData),
			)
			if err != nil {
				// 解析失败，使用原文档
				parsedDocs = append(parsedDocs, doc)
			} else {
				parsedDocs = append(parsedDocs, parsed...)
			}
		}
	}

	return parsedDocs, nil
}
