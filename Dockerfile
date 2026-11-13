# syntax=docker/dockerfile:1

# ============ Stage 1: Builder ============
FROM golang:1.25-alpine AS builder

# 国内镜像加速 (可选, CI 环境可注释掉)
ENV GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

# 先拷贝 go.mod/go.sum 做依赖缓存层
COPY go.mod go.sum ./
RUN go mod download

# 再拷贝源代码
COPY . .

# 编译
RUN go build -trimpath -ldflags="-s -w" -o /build/ad-monitor ./

# ============ Stage 2: Runtime ============
FROM alpine:3.19

# 基础运行时依赖: ca-certificates 用于 HTTPS 调用外部接口, tzdata 设置时区
RUN apk add --no-cache ca-certificates tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 拷贝二进制
COPY --from=builder /build/ad-monitor /app/ad-monitor

# 拷贝配置文件和静态资源 (config.yaml 打包进镜像, 符合你的选择)
COPY --from=builder /build/manifest /app/manifest
COPY --from=builder /build/resource /app/resource

# CSV 导出目录 (容器内通过 volume 挂载持久化)
RUN mkdir -p /app/resource/export

EXPOSE 8000

ENTRYPOINT ["/app/ad-monitor"]
