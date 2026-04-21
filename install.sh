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

if [ "$(id -u)" -eq 0 ]; then
  sudo_cmd=""
elif command -v sudo >/dev/null 2>&1; then
  sudo_cmd="sudo"
else
  echo "安装 routecheck 需要 root 或 sudo 权限" >&2
  exit 1
fi

url="https://github.com/${repo}/releases/latest/download/${file}"
if command -v curl >/dev/null 2>&1; then
  curl -fsSL -o "$tmpdir/routecheck.tar.gz" "$url"
elif command -v wget >/dev/null 2>&1; then
  wget -q -O "$tmpdir/routecheck.tar.gz" "$url"
else
  echo "需要 curl 或 wget" >&2
  exit 1
fi

tar -xf "$tmpdir/routecheck.tar.gz" -C "$tmpdir"
$sudo_cmd install -m 0755 "$tmpdir/routecheck" /usr/local/bin/routecheck

if command -v setcap >/dev/null 2>&1; then
  if $sudo_cmd setcap cap_net_raw+ep /usr/local/bin/routecheck 2>/dev/null; then
    echo "已安装 routecheck 到 /usr/local/bin/routecheck，并授予 cap_net_raw"
    routecheck "$@"
    exit 0
  fi
fi

echo "已安装 routecheck 到 /usr/local/bin/routecheck"
if [ -n "$sudo_cmd" ]; then
  echo "未能授予 cap_net_raw，正在使用 sudo 运行 routecheck" >&2
  $sudo_cmd routecheck "$@"
else
  echo "未能授予 cap_net_raw；如果遇到 raw socket 权限错误，请使用 sudo routecheck" >&2
  routecheck "$@"
fi
