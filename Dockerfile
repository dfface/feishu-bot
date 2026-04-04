# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o feishu-bot ./cmd/feishu-bot

# 运行阶段
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 复制构建产物
COPY --from=builder /app/feishu-bot /app/

# 复制配置文件示例
COPY config.yaml.example /app/config.yaml.example

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["./feishu-bot"]
