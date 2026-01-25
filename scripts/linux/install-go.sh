#!/usr/bin/env bash
set -euo pipefail

cd /tmp

# 获取最新版本号，例如 go1.23.0
LATEST_VERSION=$(curl -s https://go.dev/VERSION?m=text || true)
if [[ -z "${LATEST_VERSION:-}" ]]; then
  echo "无法自动获取最新 Go 版本号，请检查网络或稍后重试。"
  exit 1
fi

echo "检测到 Go 最新版本号：${LATEST_VERSION}"

# 让用户确认是否使用该版本
read -r -p "是否使用该版本？[Y/n] " USE_LATEST

VERSION=""
if [[ -z "${USE_LATEST}" || "${USE_LATEST}" =~ ^[Yy]$ ]]; then
  VERSION="$LATEST_VERSION"
else
  read -r -p "请输入想安装的 Go 版本（例如 go1.21.5 或 1.21.5）： " VERSION
fi
# 去掉左右空格
VERSION=$(echo "$VERSION" | xargs)
if [[ -z "${VERSION}" ]]; then
  echo "版本号不能为空，安装中止。"
  exit 1
fi
# 如果没写 go 前缀，自动补上
case "$VERSION" in
  go*)
    # 已经是 go1.25.5 这种
    ;;
  *)
    VERSION="go${VERSION}"
    ;;
esac

# 根据当前系统架构选择合适的包
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)
    GO_ARCH="amd64"
    ;;
  aarch64|arm64)
    GO_ARCH="arm64"
    ;;
  *)
    echo "未知架构：${ARCH}，请手动修改脚本中的 GO_ARCH。"
    exit 1
    ;;
esac

FILE="${VERSION}.linux-${GO_ARCH}.tar.gz"
URL="https://go.dev/dl/${FILE}"

echo "即将安装："
echo "  版本：${VERSION}"
echo "  架构：${GO_ARCH}"
echo "  下载地址：${URL}"

read -r -p "确认开始下载并安装？[Y/n] " CONFIRM
if [[ -n "${CONFIRM}" && ! "${CONFIRM}" =~ ^[Yy]$ ]]; then
  echo "用户取消安装。"
  exit 0
fi

# 下载
wget "${URL}"

# 移除旧版本
sudo rm -rf /usr/local/go

# 解压到 /usr/local
sudo tar -C /usr/local -xzf "${FILE}"

# 删除安装包（自动清理）
rm -f "${FILE}"

# 为 zsh 配置 PATH（如果尚未配置）
if ! grep -q '/usr/local/go/bin' "${HOME}/.zshrc" 2>/dev/null; then
  echo 'export PATH=$PATH:/usr/local/go/bin' >> "${HOME}/.zshrc"
  echo "已向 ~/.zshrc 追加 PATH 配置：export PATH=\$PATH:/usr/local/go/bin"
else
  echo "检测到 ~/.zshrc 中已包含 /usr/local/go/bin 到 PATH，不再重复添加。"
fi

echo
echo "Go 安装完成。请执行："
echo "  source ~/.zshrc"
echo "或重新打开一个终端，然后执行："
echo "  go version"
