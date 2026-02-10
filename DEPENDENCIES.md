# 依赖安装说明

## 新增的 Go 依赖

本次优化新增了以下依赖包，请运行 `go mod tidy` 安装：

```bash
go mod tidy
```

### 核心依赖

#### Redis客户端
```bash
go get github.com/redis/go-redis/v9
```

#### SQL解析器 (MCP服务器安全)
```bash
go get github.com/xwb1989/sqlparser
```

#### YAML配置解析
```bash
go get gopkg.in/yaml.v3
```

#### MySQL驱动
```bash
go get github.com/go-sql-driver/mysql
```

#### MCP SDK 
```bash
go get github.com/modelcontextprotocol/go-sdk
```

### 可选依赖 

#### PDF解析 (多模态增强)
```bash
go get github.com/unidoc/unipdf/v3
```

## 手动安装所有依赖

如果 `go mod tidy` 不工作，可以手动安装：

```bash
go get github.com/redis/go-redis/v9
go get github.com/xwb1989/sqlparser
go get gopkg.in/yaml.v3
go get github.com/go-sql-driver/mysql
go get github.com/modelcontextprotocol/go-sdk
```

## 验证安装

安装完成后，运行以下命令验证：

```bash
# 编译检查
go build

# 运行测试
go test ./...

# 运行性能基准测试
go test -bench=. ./test
```

## Docker 构建

如果使用Docker部署，依赖会自动在构建时安装：

```bash
docker-compose build mcp-server
docker-compose up -d
```

## 故障排查

### 问题1: go: cannot find module
**解决**: 确保在项目根目录执行命令，且 go.mod 文件存在

### 问题2: package xxx is not in GOROOT
**解决**: 运行 `go mod download` 重新下载依赖

### 问题3: 版本冲突
**解决**: 删除 go.sum 文件，重新运行 `go mod tidy`
