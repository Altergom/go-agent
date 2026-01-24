package db

import (
	"context"
	"go-agent/config"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
)

var Milvus client.Client

func NewMilvus(ctx context.Context) (client.Client, error) {
	cli, err := client.NewClient(ctx, client.Config{
		Address:  config.Cfg.MilvusConf.MilvusAddr,
		Username: config.Cfg.MilvusConf.MilvusUserName,
		Password: config.Cfg.MilvusConf.MilvusPassword,
	})
	if err != nil {
		return nil, err
	}

	return cli, nil
}
