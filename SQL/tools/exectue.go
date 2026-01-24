package tools

import (
	"context"
	"fmt"
	"go-agent/config"
	"os"
	"os/exec"

	"github.com/cloudwego/eino-ext/components/tool/mcp/officialmcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func SQLExecute(ctx context.Context, sql string) (string, error) {
	mcpTool, err := GetMCPTool(ctx)
	if err != nil {
		return "", err
	}

	var targetTool tool.InvokableTool
	for _, t := range mcpTool {
		if invokable, ok := t.(tool.InvokableTool); ok {
			info, _ := invokable.Info(ctx)
			if info.Name == "MySQL" {
				targetTool = invokable
				break
			}
		}
	}

	if targetTool == nil {
		return "", fmt.Errorf("未找到指定的数据库执行工具")
	}

	args := fmt.Sprintf(`{"sql": "%s"}`, sql)

	// 使用InvokableRun通过MCP协议发送请求到Server
	result, err := targetTool.InvokableRun(ctx, args)
	if err != nil {
		return "", fmt.Errorf("MCP 工具执行失败: %w", err)
	}

	return result, nil
}

func GetMCPTool(ctx context.Context) ([]tool.BaseTool, error) {
	cli := mcp.NewClient(&mcp.Implementation{
		Name:    "go-agent-client",
		Version: "1.0.0",
	}, nil)

	cmd := exec.Command("npx", "-y", "mcp-server-mysql")

	cmd.Env = append(os.Environ(),
		"MYSQL_HOST="+config.Cfg.MySQLConf.Host,
		"MYSQL_PORT="+config.Cfg.MySQLConf.Port,
		"MYSQL_USER="+config.Cfg.MySQLConf.Username,
		"MYSQL_PASS="+config.Cfg.MySQLConf.Password,
		"MYSQL_DB="+config.Cfg.MySQLConf.Database,
	)

	transport := &mcp.CommandTransport{
		Command: cmd,
	}

	session, err := cli.Connect(ctx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("MCP 客户端连接失败: %w", err)
	}

	return officialmcp.GetTools(ctx, &officialmcp.Config{
		Cli:          session,
		ToolNameList: []string{},
	})
}
