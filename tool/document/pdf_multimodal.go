package document

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/cloudwego/eino/components/document/parser"
	"github.com/cloudwego/eino/schema"
)

// MultimodalPDFParser 多模态PDF解析器
type MultimodalPDFParser struct {
	textParser  parser.Parser
	tableParser TableExtractor
	imageParser ImageExtractor
}

// TableExtractor 表格提取接口
type TableExtractor interface {
	ExtractTables(ctx context.Context, content []byte) ([]*TableData, error)
}

// ImageExtractor 图片提取接口
type ImageExtractor interface {
	ExtractImages(ctx context.Context, content []byte) ([]*ImageData, error)
}

// TableData 表格数据
type TableData struct {
	Content  string             // Markdown格式的表格
	Page     int                // 所在页码
	Position map[string]float64 // 位置信息 {x, y, width, height}
}

// ImageData 图片数据
type ImageData struct {
	Base64   string             // Base64编码的图片
	Format   string             // 图片格式 (png, jpg等)
	Page     int                // 所在页码
	Position map[string]float64 // 位置信息
}

// NewMultimodalPDFParser 创建多模态PDF解析器
func NewMultimodalPDFParser(ctx context.Context) (*MultimodalPDFParser, error) {
	textParser := parser.TextParser{}

	return &MultimodalPDFParser{
		textParser:  textParser,
		tableParser: &SimpleTableExtractor{},
		imageParser: &SimpleImageExtractor{},
	}, nil
}

// Parse 解析PDF文档
func (p *MultimodalPDFParser) Parse(ctx context.Context, content io.Reader) ([]*schema.Document, error) {
	// 读取全部内容
	data, err := io.ReadAll(content)
	if err != nil {
		return nil, fmt.Errorf("读取内容失败: %w", err)
	}

	documents := make([]*schema.Document, 0)

	// 提取文本块
	textDocs, err := p.extractText(ctx, data)
	if err == nil {
		documents = append(documents, textDocs...)
	}

	// 提取表格
	tables, err := p.tableParser.ExtractTables(ctx, data)
	if err == nil {
		for _, table := range tables {
			doc := &schema.Document{
				ID:      fmt.Sprintf("table_page_%d", table.Page),
				Content: table.Content,
				MetaData: map[string]interface{}{
					"type":     "table",
					"page":     table.Page,
					"position": table.Position,
				},
			}
			documents = append(documents, doc)
		}
	}

	// 提取图片
	images, err := p.imageParser.ExtractImages(ctx, data)
	if err == nil {
		for i, image := range images {
			doc := &schema.Document{
				ID:      fmt.Sprintf("image_%d_page_%d", i, image.Page),
				Content: fmt.Sprintf("[Image on page %d]", image.Page),
				MetaData: map[string]interface{}{
					"type":     "image",
					"base64":   image.Base64,
					"format":   image.Format,
					"page":     image.Page,
					"position": image.Position,
				},
			}
			documents = append(documents, doc)
		}
	}

	return documents, nil
}

// extractText 提取文本内容
func (p *MultimodalPDFParser) extractText(ctx context.Context, data []byte) ([]*schema.Document, error) {
	// 使用基础文本解析器
	reader := strings.NewReader(string(data))
	return p.textParser.Parse(ctx, reader)
}

// SimpleTableExtractor 简单表格提取器（占位实现）
type SimpleTableExtractor struct{}

func (e *SimpleTableExtractor) ExtractTables(ctx context.Context, content []byte) ([]*TableData, error) {
	// unipdf实现
	return nil, nil
}

// SimpleImageExtractor 简单图片提取器（占位实现）
type SimpleImageExtractor struct{}

func (e *SimpleImageExtractor) ExtractImages(ctx context.Context, content []byte) ([]*ImageData, error) {
	// unipdf实现
	return nil, nil
}

// GeminiTableExtractor Gemini Vision表格提取器
type ArkTableExtractor struct {
	// geminiClient vision.Client // 需要集成Gemini Vision API
}

func (a *ArkTableExtractor) ExtractTables(ctx context.Context, content []byte) ([]*TableData, error) {
	// Vision模型实现
	// 1. 将PDF页面转为图片
	// 2. Ark Vision API识别表格结构
	// 3. 将识别结果转换为Markdown格式
	return nil, fmt.Errorf("Ark表格提取器未实现")
}

// Base64EncodeImage 将图片编码为Base64
func Base64EncodeImage(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64DecodeImage 从Base64解码图片
func Base64DecodeImage(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}
