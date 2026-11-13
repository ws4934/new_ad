#!/bin/sh
# 运行 data 卷里的二进制 (在线更新替换的是它)。
# 镜像内种子变化 (拉新镜像 / 回退旧镜像 / 本地重新构建) 时用种子覆盖; 同一镜像重启则保留在线更新的结果。
set -eu

seed=/app/ad-monitor
bin=/app/data/bin/ad-monitor
mkdir -p "$(dirname "$bin")"

sum=$(sha256sum "$seed" | cut -d' ' -f1)
if [ ! -x "$bin" ] || [ "$(cat "$bin.seed" 2>/dev/null)" != "$sum" ]; then
    cp "$seed" "$bin.tmp"
    chmod 755 "$bin.tmp"
    mv "$bin.tmp" "$bin"
    echo "$sum" > "$bin.seed"
fi

exec "$bin" "$@"
