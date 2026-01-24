package tools

import (
	"context"
	"encoding/json"
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
	var toolNames []string
	for _, t := range mcpTool {
		if invokable, ok := t.(tool.InvokableTool); ok {
			info, _ := invokable.Info(ctx)
			toolNames = append(toolNames, info.Name)
			if info.Name == "mysql_query" {
				targetTool = invokable
				break
			}
		}
	}

	if targetTool == nil {
		return "", fmt.Errorf("未找到指定的数据库执行工具, 当前可用工具: %v", toolNames)
	}

	// 使用 json.Marshal 自动处理 SQL 中的换行符和特殊字符转义
	params := map[string]string{"sql": sql}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("序列化 MCP 参数失败: %w", err)
	}

	// 使用InvokableRun通过MCP协议发送请求到Server
	result, err := targetTool.InvokableRun(ctx, string(paramsJSON))
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
		ToolNameList: []string{"mysql_query", "list_tables", "describe_table"},
	})
}
