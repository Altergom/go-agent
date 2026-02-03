package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-agent/config"

	"github.com/elastic/go-elasticsearch/v8"
)

var ES *elasticsearch.Client

func NewES() (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: config.Cfg.ESConf.Addresses,
		Username:  config.Cfg.ESConf.Username,
		Password:  config.Cfg.ESConf.Password,
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	// 自动初始化索引映射
	ctx := context.Background()
	indexName := config.Cfg.ESConf.Index

	// 检查索引是否存在
	res, err := client.Indices.Exists([]string{indexName})
	if err != nil {
		return nil, fmt.Errorf("检查 ES 索引失败: %w", err)
	}

	if res.StatusCode == 404 {
		// 索引不存在，开始创建并定义 Mapping
		// 定义 Mapping 结构
		mapping := map[string]interface{}{
			"mappings": map[string]interface{}{
				"properties": map[string]interface{}{
					"content":  map[string]interface{}{"type": "text"},
					"metadata": map[string]interface{}{"type": "object"},
					"content_vector": map[string]interface{}{
						"type":       "dense_vector",
						"dims":       2560,
						"index":      true,
						"similarity": "cosine",
					},
				},
			},
		}

		body, _ := json.Marshal(mapping)
		createRes, err := client.Indices.Create(
			indexName,
			client.Indices.Create.WithContext(ctx),
			client.Indices.Create.WithBody(bytes.NewReader(body)),
		)
		if err != nil {
			return nil, fmt.Errorf("创建 ES 索引失败: %w", err)
		}
		if createRes.IsError() {
			return nil, fmt.Errorf("创建 ES 索引返回错误: %s", createRes.String())
		}
		fmt.Printf("ES 索引 [%s] 初始化成功\n", indexName)
	}

	return client, nil
}
