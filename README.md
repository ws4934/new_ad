# ad-monitor

广告监测上报平台，附带淘宝客超级红包直发状态监控。

## 安装（Docker）

服务器需已安装 Docker 与 docker compose 插件，仅支持 linux/amd64。

```bash
curl -fsSL https://raw.githubusercontent.com/ws4934/new_ad/main/install.sh | sudo bash
```

- 默认安装到 `/opt/ad-monitor`，端口 `8090`，安装完成后会打印后台密码。
- 可选环境变量：`INSTALL_DIR`、`APP_PORT`、`BASE_URL`（对外域名）、`VERSION`（如 `v1.0.0`，默认最新版）。

```bash
curl -fsSL https://raw.githubusercontent.com/ws4934/new_ad/main/install.sh \
  | sudo BASE_URL=https://ad.example.com APP_PORT=8090 bash
```

配置文件为 `/opt/ad-monitor/manifest/config/config.yaml`，修改后执行：

```bash
cd /opt/ad-monitor && docker compose restart app
```

## 更新

任选其一：

- 后台「系统更新」页面点击「立即更新」，程序自动下载新版本并重启。
- 重新执行上面的安装命令，会拉取最新镜像并重启，不会覆盖 `.env` 和 `config.yaml`。

## 发布新版本

```bash
git tag v1.0.0 && git push origin v1.0.0
```

推送 tag 后 GitHub Actions 会构建二进制与镜像（`ghcr.io/ws4934/new_ad`）并发布 Release。首次发布后需在 GitHub 的 Packages 设置中把镜像设为 Public，否则服务器需要先 `docker login ghcr.io`。
