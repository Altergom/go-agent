package trace

import (
	"context"
	"log"
	"os"

	ccb "github.com/cloudwego/eino-ext/callbacks/cozeloop"
	"github.com/cloudwego/eino/callbacks"
	"github.com/coze-dev/cozeloop-go"
)

func NewCozeLoop(ctx context.Context) (func(), error) {
	// 检查官方要求的环境变量是否存在
	if os.Getenv("COZELOOP_JWT_OAUTH_CLIENT_ID") == "" {
		log.Println("CozeLoop OAuth 配置缺失，跳过初始化")
		return func() {}, nil
	}

	// NewClient会自动读取环境变量:
	client, err := cozeloop.NewClient()
	if err != nil {
		return nil, err
	}

	handler := ccb.NewLoopHandler(client)
	callbacks.AppendGlobalHandlers(handler)
	log.Println("CozeLoop 全局回调已启用 (OAuth 模式)")

	return func() {
		client.Close(ctx)
	}, nil
}
