# syntax=docker/dockerfile:1

# ============ Stage 1: Builder ============
FROM golang:1.25-alpine AS builder

# 国内镜像加速 (CI 通过 --build-arg GOPROXY=https://proxy.golang.org,direct 覆盖)
ARG GOPROXY=https://goproxy.cn,direct
# 版本号由 CI 注入 (git tag), 本地构建为 dev (dev 版本不支持在线更新)
ARG VERSION=dev
ENV GOPROXY=${GOPROXY} \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

# 先拷贝 go.mod/go.sum (+ replace 指向的本地 SDK) 做依赖缓存层
COPY go.mod go.sum ./
COPY third_party ./third_party
RUN go mod download

# 再拷贝源代码
COPY . .

# 编译 (前端静态文件已 embed 进二进制)
RUN go build -trimpath -ldflags="-s -w -X ad-monitor/internal/logic/updater.Version=${VERSION}" -o /build/ad-monitor ./

# ============ Stage 2: Runtime ============
FROM alpine:3.19

# 基础运行时依赖: ca-certificates 用于 HTTPS 调用外部接口, tzdata 设置时区
RUN apk add --no-cache ca-certificates tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 镜像自带的二进制只作为"种子": entrypoint 把它复制到 data 卷后运行,
# 在线更新替换的是卷里的那份, 重建容器不会丢; 拉到更新的镜像时种子会覆盖旧版本。
COPY --from=builder /build/ad-monitor /app/ad-monitor
COPY docker-entrypoint.sh /app/docker-entrypoint.sh

# 默认配置 (docker-compose.prod.yml 会用宿主机的 config.yaml 覆盖)
COPY --from=builder /build/manifest /app/manifest

# CSV 导出目录 (容器内通过 volume 挂载持久化)
RUN chmod 755 /app/docker-entrypoint.sh && mkdir -p /app/resource/export /app/data/bin

EXPOSE 8000

ENTRYPOINT ["/app/docker-entrypoint.sh"]
