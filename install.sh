#!/bin/sh
set -eu

repo="${ROUTECHECK_REPO:-libai07/routecheck}"
arch="$(uname -m)"

case "$arch" in
  x86_64|amd64)
    file="routecheck-linux-amd64.tar.gz"
    ;;
  aarch64|arm64)
    file="routecheck-linux-arm64.tar.gz"
    ;;
  *)
    echo "不支持的架构: $arch" >&2
    exit 1
    ;;
esac

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT INT TERM
PATH="/usr/local/bin:$PATH"
export PATH

download() {
  url="$1"
  dest="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "$dest" "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -q -O "$dest" "$url"
  else
    echo "需要 curl 或 wget" >&2
    exit 1
  fi
}

if [ "$(id -u)" -eq 0 ]; then
  sudo_cmd=""
elif command -v sudo >/dev/null 2>&1; then
  sudo_cmd="sudo"
else
  echo "安装 routecheck 需要 root 或 sudo 权限" >&2
  exit 1
fi

url="https://github.com/${repo}/releases/latest/download/${file}"
download "$url" "$tmpdir/routecheck.tar.gz"

tar -xf "$tmpdir/routecheck.tar.gz" -C "$tmpdir"
$sudo_cmd install -m 0755 "$tmpdir/routecheck" /usr/local/bin/routecheck

echo "已安装 routecheck 到 /usr/local/bin/routecheck"

if ! command -v nexttrace >/dev/null 2>&1 && ! command -v nexttrace-tiny >/dev/null 2>&1; then
  if ! command -v bash >/dev/null 2>&1; then
    echo "未找到 nexttrace 或 nexttrace-tiny，且系统缺少 bash，无法自动安装 NextTrace" >&2
    exit 1
  fi

  echo "未找到 nexttrace 或 nexttrace-tiny，正在安装 NextTrace..."
  download "https://nxtrace.org/nt" "$tmpdir/nexttrace-install.sh"
  if [ -n "$sudo_cmd" ]; then
    $sudo_cmd bash "$tmpdir/nexttrace-install.sh"
  else
    bash "$tmpdir/nexttrace-install.sh"
  fi

  if ! command -v nexttrace >/dev/null 2>&1 && ! command -v nexttrace-tiny >/dev/null 2>&1; then
    echo "NextTrace 安装后仍未在 PATH 中找到 nexttrace 或 nexttrace-tiny" >&2
    exit 1
  fi
fi

/usr/local/bin/routecheck "$@"
