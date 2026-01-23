package trace

import (
	"go-agent/config"
	"log"

	"github.com/cloudwego/eino-ext/callbacks/langsmith"
	"github.com/cloudwego/eino/callbacks"
)

func NewLangSmith() error {
	traceHandler, err := langsmith.NewLangsmithHandler(&langsmith.Config{
		APIKey: config.Cfg.LangSmithConf.APIKey,
		APIURL: config.Cfg.LangSmithConf.APIUrl,
	})
	if err != nil {
		return err
	}
	callbacks.AppendGlobalHandlers(traceHandler)
	log.Println("LangSmith 全局回调已启用")

	return nil
}
