#!/usr/bin/env bash
# ad-monitor 一键安装 / 更新 (Docker)
#
#   curl -fsSL https://raw.githubusercontent.com/ws4934/new_ad/main/install.sh | sudo bash
#   # 国内服务器连不上 GitHub 时走代理:
#   curl -fsSL https://gh-proxy.com/https://raw.githubusercontent.com/ws4934/new_ad/main/install.sh | sudo bash
#
# 可选环境变量:
#   GH_PROXY     GitHub 代理前缀, 如 https://gh-proxy.com/ ; 不设置则自动检测, 设为空串表示强制直连
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

# GitHub 访问: 直连不通时依次尝试公共代理 (代理可看到下载内容)
gh_ok() { curl -fsS --connect-timeout 5 -m 20 -o /dev/null "$1https://raw.githubusercontent.com/${REPO}/main/install.sh" 2>/dev/null; }
if [ -z "${GH_PROXY+x}" ]; then
    GH_PROXY=""
    if ! gh_ok ""; then
        for p in https://gh-proxy.com/ https://ghproxy.net/ https://ghfast.top/; do
            if gh_ok "$p"; then GH_PROXY=$p; break; fi
        done
        [ -n "$GH_PROXY" ] || die "无法访问 GitHub, 也没有可用的代理; 可手动指定: GH_PROXY=https://<代理>/ bash install.sh"
    fi
fi
[ -n "$GH_PROXY" ] && info "GitHub 访问走代理: ${GH_PROXY}"

# 版本: 默认最新 Release (从 SHA256SUMS 解析, 不依赖 api.github.com), 取不到时用 main 分支文件 + latest 镜像
if [ -z "${VERSION:-}" ]; then
    VERSION=$(curl -fsSL --connect-timeout 5 -m 30 "${GH_PROXY}https://github.com/${REPO}/releases/latest/download/SHA256SUMS" 2>/dev/null \
        | sed -n 's/.*ad-monitor_\(v[0-9.]*\)_linux_amd64\.tar\.gz.*/\1/p' | head -n1 || true)
fi
REF="${VERSION:-main}"
TAG="${VERSION:-latest}"
RAW="${RAW:-${GH_PROXY}https://raw.githubusercontent.com/${REPO}/${REF}}"
info "版本: ${TAG}  目录: ${INSTALL_DIR}"

mkdir -p "${INSTALL_DIR}"/{manifest/config,hack/sql,data}
cd "${INSTALL_DIR}"

fetch() { curl -fsSL --connect-timeout 10 -m 120 "${RAW}/$1" -o "$2.tmp" && mv "$2.tmp" "$2" || die "下载 $1 失败"; }

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

CONFIG_CREATED=0
if [ ! -f manifest/config/config.yaml ]; then
    CONFIG_CREATED=1
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
# 走了代理时, 程序内在线更新也用同一代理
if [ -n "$GH_PROXY" ] && grep -q 'githubProxy: ""' manifest/config/config.yaml; then
    sed -i "s#githubProxy: \"\"#githubProxy: \"${GH_PROXY}\"#" manifest/config/config.yaml
fi

# 拉镜像: 原地址失败 (国内常见) 时依次尝试国内镜像源, 成功后 tag 回原名, compose 无需改动
pull() {
    local img=$1; shift
    info "拉取镜像 ${img}"
    timeout 300 docker pull -q "$img" >/dev/null 2>&1 && return 0
    for m in "$@"; do
        info "  原地址失败, 尝试镜像源 ${m}"
        if timeout 600 docker pull -q "$m" >/dev/null 2>&1 && docker tag "$m" "$img"; then return 0; fi
    done
    return 1
}
have() { docker image inspect "$1" >/dev/null 2>&1; }

pull "${IMAGE}:${TAG}" "ghcr.nju.edu.cn/${REPO}:${TAG}" || die "拉取 ${IMAGE}:${TAG} 失败, 请检查网络或配置 Docker 镜像加速"
for img in postgres:17-alpine redis:7-alpine; do
    have "$img" || pull "$img" "docker.1ms.run/library/${img}" "docker.m.daocloud.io/library/${img}" \
        || die "拉取 ${img} 失败, 请为 Docker 配置镜像加速 (如阿里云容器镜像服务的加速器地址)"
done
docker compose up -d --no-build --remove-orphans

PORT=$(sed -n 's/^APP_PORT=//p' .env)
info "完成, 访问 http://<服务器IP>:${PORT:-8090}"
if [ "${CONFIG_CREATED}" = 1 ]; then
    info "后台密码: ${ADMIN_PASSWORD}  (保存在 ${INSTALL_DIR}/manifest/config/config.yaml)"
fi
[ -z "${BASE_URL:-}" ] && [ "${FIRST_INSTALL}" = 1 ] && \
    info "提示: 请在 config.yaml 中修改 callbackBaseUrl / monitorBaseUrl 为你的域名, 然后 docker compose restart app"
info "查看日志: cd ${INSTALL_DIR} && docker compose logs -f app"
