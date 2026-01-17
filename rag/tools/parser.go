package tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/document/parser/html"
	"github.com/cloudwego/eino-ext/components/document/parser/pdf"
	"github.com/cloudwego/eino/components/document/parser"
)

// Parser 解析器 解析文档为[]*schema.Document数据结构
var Parser parser.Parser

func NewParser(ctx context.Context) (parser.Parser, error) {
	textParser := parser.TextParser{}

	htmlParser, err := html.NewParser(ctx, &html.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create HTML parser: %w", err)
	}

	pdfParser, err := pdf.NewPDFParser(ctx, &pdf.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF parser: %w", err)
	}

	extParser, err := parser.NewExtParser(ctx, &parser.ExtParserConfig{
		Parsers: map[string]parser.Parser{
			".html": htmlParser,
			".htm":  htmlParser,
			".pdf":  pdfParser,
			".txt":  textParser,
			".md":   textParser,
		},
		FallbackParser: textParser,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ext parser: %w", err)
	}

	return extParser, nil
}
