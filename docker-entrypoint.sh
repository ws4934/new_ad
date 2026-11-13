#!/bin/sh
# 运行 data 卷里的二进制 (在线更新替换的是它), 首次启动或镜像更新时用镜像内的种子覆盖。
set -eu

seed=/app/ad-monitor
bin=/app/data/bin/ad-monitor
mkdir -p "$(dirname "$bin")"

if [ ! -x "$bin" ] || "$seed" --newer-than "$("$bin" --version 2>/dev/null || echo none)"; then
    cp "$seed" "$bin.tmp"
    chmod 755 "$bin.tmp"
    mv "$bin.tmp" "$bin"
fi

exec "$bin" "$@"
