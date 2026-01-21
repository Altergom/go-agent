package db

import (
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

	return client, nil
}
