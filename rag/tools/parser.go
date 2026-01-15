package tools

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/parser/html"
	"github.com/cloudwego/eino-ext/components/document/parser/pdf"
	"github.com/cloudwego/eino/components/document/parser"
)

type Parser struct {
	PDFParser  *pdf.PDFParser
	HTMLParser *html.Parser
	ExtParser  *parser.ExtParser
}

var Par *Parser

func NewParser(ctx context.Context) (*Parser, error) {
	textParser := parser.TextParser{}
	htmlParser, err := html.NewParser(ctx, &html.Config{})
	if err != nil {
		return nil, err
	}

	pdfParser, err := pdf.NewPDFParser(ctx, &pdf.Config{})
	if err != nil {
		return nil, err
	}

	// 创建扩展解析器
	extParser, err := parser.NewExtParser(ctx, &parser.ExtParserConfig{
		// 注册特定扩展名的解析器
		Parsers: map[string]parser.Parser{
			".html": htmlParser,
			".pdf":  pdfParser,
		},
		// 设置默认解析器，用于处理未知格式
		FallbackParser: textParser,
	})
	if err != nil {
		return nil, err
	}
	p := &Parser{
		PDFParser:  pdfParser,
		HTMLParser: htmlParser,
		ExtParser:  extParser,
	}

	return p, nil
}
