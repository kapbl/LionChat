# 使用官方Go镜像作为构建环境
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/app

# 使用轻量级的alpine镜像作为运行环境
FROM alpine:latest

# 安装ca-certificates用于HTTPS请求
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 复制配置文件
COPY --from=builder /app/config ./config

# 复制静态资源
COPY --from=builder /app/static ./static
COPY --from=builder /app/resources ./resources

# 创建日志目录
RUN mkdir -p /root/logs

# 设置文件权限
RUN chown -R appuser:appgroup /root

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8081

# 设置环境变量
ENV GIN_MODE=release
ENV CONFIG_PATH=./config/config.yaml

# 运行应用
CMD ["./main"]