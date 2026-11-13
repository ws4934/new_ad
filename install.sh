#!/usr/bin/env bash
# ad-monitor 一键安装 / 更新 (Docker)
#
#   curl -fsSL https://raw.githubusercontent.com/ws4934/new_ad/main/install.sh | sudo bash
#
# 可选环境变量:
#   INSTALL_DIR  安装目录, 默认 /opt/ad-monitor
#   APP_PORT     对外端口, 默认 8090 (仅首次安装生效)
#   BASE_URL     对外访问域名, 如 https://ad.example.com (仅首次安装写入 config.yaml)
#   VERSION      指定版本, 如 v1.0.0, 默认最新 Release
#
# 再次运行即为更新: 拉取最新镜像并重启, 不会覆盖 .env 与 config.yaml。
set -euo pipefail

REPO="ws4934/new_ad"
IMAGE="ghcr.io/${REPO}"
INSTALL_DIR="${INSTALL_DIR:-/opt/ad-monitor}"

info() { printf '\033[32m[ad-monitor]\033[0m %s\n' "$*"; }
die()  { printf '\033[31m[ad-monitor]\033[0m %s\n' "$*" >&2; exit 1; }
rand() { head -c 1024 /dev/urandom | LC_ALL=C tr -dc 'A-Za-z0-9' | cut -c "1-${1:-24}"; }

command -v curl >/dev/null || die "缺少 curl"
command -v docker >/dev/null || die "未安装 Docker, 请先安装: curl -fsSL https://get.docker.com | sh"
docker compose version >/dev/null 2>&1 || die "缺少 docker compose 插件"

# 版本: 默认最新 Release, 没有 Release 时用 main 分支文件 + latest 镜像
if [ -z "${VERSION:-}" ]; then
    VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null \
        | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n1 || true)
fi
REF="${VERSION:-main}"
TAG="${VERSION:-latest}"
RAW="${RAW:-https://raw.githubusercontent.com/${REPO}/${REF}}"
info "版本: ${TAG}  目录: ${INSTALL_DIR}"

mkdir -p "${INSTALL_DIR}"/{manifest/config,hack/sql,data}
cd "${INSTALL_DIR}"

fetch() { curl -fsSL "${RAW}/$1" -o "$2.tmp" && mv "$2.tmp" "$2" || die "下载 $1 失败"; }

# 每次都更新的文件
fetch docker-compose.prod.yml docker-compose.yml
fetch hack/sql/init.sql hack/sql/init.sql

# 首次安装: 生成 .env 与 config.yaml (随机数据库密码 / 后台密码 / JWT 密钥)
FIRST_INSTALL=0
if [ ! -f .env ]; then
    FIRST_INSTALL=1
    DB_PASSWORD=$(rand 24)
    cat > .env <<EOF
DB_USER=postgres
DB_PASSWORD=${DB_PASSWORD}
DB_NAME=ad_monitor
APP_PORT=${APP_PORT:-8090}
APP_TAG=${TAG}
EOF
    chmod 600 .env
else
    # 更新: 只改镜像版本
    if grep -q '^APP_TAG=' .env; then
        sed -i "s#^APP_TAG=.*#APP_TAG=${TAG}#" .env
    else
        echo "APP_TAG=${TAG}" >> .env
    fi
fi

if [ ! -f manifest/config/config.yaml ]; then
    fetch manifest/config/config.yaml manifest/config/config.yaml
    DB_PASSWORD=$(sed -n 's/^DB_PASSWORD=//p' .env)
    ADMIN_PASSWORD=$(rand 16)
    sed -i \
        -e "s#pgsql:postgres:[^@]*@tcp(127.0.0.1:5432)#pgsql:postgres:${DB_PASSWORD}@tcp(postgres:5432)#" \
        -e 's#"127.0.0.1:6379"#"redis:6379"#g' \
        -e "s#adminPassword: \"[^\"]*\"#adminPassword: \"${ADMIN_PASSWORD}\"#" \
        -e "s#jwtSecret: \"[^\"]*\"#jwtSecret: \"$(rand 32)\"#" \
        manifest/config/config.yaml
    if [ -n "${BASE_URL:-}" ]; then
        sed -i \
            -e "s#callbackBaseUrl: \"[^\"]*\"#callbackBaseUrl: \"${BASE_URL}\"#" \
            -e "s#monitorBaseUrl: \"[^\"]*\"#monitorBaseUrl: \"${BASE_URL}\"#" \
            manifest/config/config.yaml
    fi
    chmod 600 manifest/config/config.yaml
fi

info "拉取镜像 ${IMAGE}:${TAG}"
docker compose pull app || die "拉取镜像失败 (若 ghcr 包为私有, 请先 docker login ghcr.io 或在 GitHub 将包设为 Public)"
docker compose up -d --no-build --remove-orphans

PORT=$(sed -n 's/^APP_PORT=//p' .env)
info "完成, 访问 http://<服务器IP>:${PORT:-8090}"
if [ "${FIRST_INSTALL}" = 1 ] && [ -n "${ADMIN_PASSWORD:-}" ]; then
    info "后台密码: ${ADMIN_PASSWORD}  (保存在 ${INSTALL_DIR}/manifest/config/config.yaml)"
fi
[ -z "${BASE_URL:-}" ] && [ "${FIRST_INSTALL}" = 1 ] && \
    info "提示: 请在 config.yaml 中修改 callbackBaseUrl / monitorBaseUrl 为你的域名, 然后 docker compose restart app"
info "查看日志: cd ${INSTALL_DIR} && docker compose logs -f app"
