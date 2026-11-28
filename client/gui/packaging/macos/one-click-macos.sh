#!/usr/bin/env bash
# 一键在 macOS 上构建 GOSTC v2.0.13 图形版 universal .dmg 并上传到面板下载目录。
#
# 用法（在已 clone 的仓库里运行）：
#   git clone https://github.com/PonChunhui/gostc-open.git
#   cd gostc-open
#   bash client/gui/packaging/macos/one-click-macos.sh
#
# 依赖：macOS 11+、Xcode Command Line Tools（clang/lipo/sips/iconutil/hdiutil/codesign）、Go>=1.24。
# 产物：client/gui/dist/GOSTC_2.0.13_macos_universal.dmg，并上传为 downloads/gostc-macos-universal.dmg。
set -euo pipefail

VER="${1:-v2.0.13}"
PANEL_HOST="${PANEL_HOST:-124.222.160.204}"
PANEL_USER="${PANEL_USER:-root}"
PANEL_PASS="${PANEL_PASS:-idc.linux66.CN}"
DOWNLOAD_DIR="/usr/local/gostc-admin/downloads"

[ "$(uname)" = "Darwin" ] || { echo "错误：必须在 macOS 上运行（GUI 依赖 Cocoa/WebKit cgo）" >&2; exit 1; }
command -v go >/dev/null 2>&1 || { echo "错误：未找到 go，请先安装 Go>=1.24" >&2; exit 1; }
xcode-select -p >/dev/null 2>&1 || { echo "错误：缺 Xcode CLT，先运行 xcode-select --install" >&2; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CLIENT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"   # client/
cd "$CLIENT_DIR"

echo "==> 构建 universal .dmg ($VER)"
bash gui/packaging/macos/package.sh "$VER"
DMG="gui/dist/GOSTC_${VER#v}_macos_universal.dmg"
[ -f "$DMG" ] || { echo "错误：未产出 $DMG" >&2; exit 1; }
echo "==> 产物: $DMG ($(du -h "$DMG" | cut -f1))"

echo "==> 上传到面板 $PANEL_HOST"
ASK="$(mktemp)"
printf '#!/bin/sh\necho "%s"\n' "$PANEL_PASS" > "$ASK"
chmod 700 "$ASK"
export SSH_ASKPASS="$ASK" SSH_ASKPASS_REQUIRE=force DISPLAY=localhost:0
SSH_OPTS="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
scp $SSH_OPTS "$DMG" "$PANEL_USER@$PANEL_HOST:$DOWNLOAD_DIR/gostc-macos-universal.dmg"
# 移除旧的 arm64 单架构包，避免下载页残留旧版
ssh $SSH_OPTS "$PANEL_USER@$PANEL_HOST" "rm -f $DOWNLOAD_DIR/gostc-macos-arm64.dmg"
rm -f "$ASK"

echo "完成：已上传 gostc-macos-universal.dmg ($VER)，下载页即可获取。"
