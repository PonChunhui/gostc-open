#!/usr/bin/env bash
# 在 macOS 上构建 GOSTC GUI 的 .app / .dmg 安装包（universal 二进制：amd64 + arm64）。
#
# 依赖：macOS 11+、Xcode Command Line Tools（clang / sips / iconutil / hdiutil / codesign）、Go >= 1.24
# 用法：bash package.sh [版本号]        # 版本号默认 dev，CI 中传入 tag 名（自动去掉前缀 v）
# 产物：client/gui/dist/GOSTC_<版本>_macos_universal.dmg
#
# 说明：
# - GUI 依赖 cgo（WKWebView / 托盘），只能在 macOS 上编译，无法从 Windows/Linux 交叉产出 .app。
# - 前端资源已内嵌（client/webui/backend/web/dist.zip），无需 Node.js。
# - 未配置 Apple 开发者证书时仅做 ad-hoc 签名；分发后首次打开需右键“打开”或解除 quarantine，
#   详见 client/gui/README.md。
set -euo pipefail

GUI_DIR="$(cd "$(dirname "$0")/../.." && pwd)" # client/gui
CLIENT_DIR="$(dirname "$GUI_DIR")"            # client（go 模块根目录）
VERSION="${1:-dev}"
VERSION="${VERSION#v}"
OUT_DIR="${OUT_DIR:-$GUI_DIR/dist}"
APP_NAME="GOSTC"
BIN_NAME="gostc-gui"

if [ "$(uname)" != "Darwin" ]; then
  echo "错误：必须在 macOS 上运行（GUI 依赖 Cocoa/WebKit，需要 cgo）" >&2
  exit 1
fi
command -v go >/dev/null 2>&1 || { echo "错误：未找到 go" >&2; exit 1; }

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT
mkdir -p "$OUT_DIR"

cd "$CLIENT_DIR"
export CGO_ENABLED=1
export GOOS=darwin

echo "==> 编译本机架构（$(uname -m)）"
if [ "$(uname -m)" = "arm64" ]; then
  GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o "$TMP_DIR/$BIN_NAME-arm64" ./gui/
  echo "==> 交叉编译 amd64（clang -arch x86_64）"
  GOARCH=amd64 CC="clang -arch x86_64" go build -trimpath -ldflags "-s -w" -o "$TMP_DIR/$BIN_NAME-amd64" ./gui/
else
  GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o "$TMP_DIR/$BIN_NAME-amd64" ./gui/
  echo "==> 交叉编译 arm64（clang -arch arm64）"
  GOARCH=arm64 CC="clang -arch arm64" go build -trimpath -ldflags "-s -w" -o "$TMP_DIR/$BIN_NAME-arm64" ./gui/
fi

echo "==> 合并 universal 二进制"
lipo -create -output "$TMP_DIR/$BIN_NAME" "$TMP_DIR/$BIN_NAME-arm64" "$TMP_DIR/$BIN_NAME-amd64"

echo "==> 组装 .app"
APP="$TMP_DIR/$APP_NAME.app"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"
cp "$TMP_DIR/$BIN_NAME" "$APP/Contents/MacOS/$BIN_NAME"
chmod +x "$APP/Contents/MacOS/$BIN_NAME"

# 图标：logo.png(128x128) -> iconset -> icns（大尺寸由 sips 放大，后续可替换高清源图）
SRC_ICON="$GUI_DIR/static/logo.png"
ICONSET="$TMP_DIR/AppIcon.iconset"
mkdir -p "$ICONSET"
sips -z 16 16     "$SRC_ICON" --out "$ICONSET/icon_16x16.png"      >/dev/null
sips -z 32 32     "$SRC_ICON" --out "$ICONSET/icon_16x16@2x.png"   >/dev/null
sips -z 32 32     "$SRC_ICON" --out "$ICONSET/icon_32x32.png"      >/dev/null
sips -z 64 64     "$SRC_ICON" --out "$ICONSET/icon_32x32@2x.png"   >/dev/null
sips -z 128 128   "$SRC_ICON" --out "$ICONSET/icon_128x128.png"    >/dev/null
sips -z 256 256   "$SRC_ICON" --out "$ICONSET/icon_128x128@2x.png" >/dev/null
sips -z 256 256   "$SRC_ICON" --out "$ICONSET/icon_256x256.png"    >/dev/null
sips -z 512 512   "$SRC_ICON" --out "$ICONSET/icon_256x256@2x.png" >/dev/null
sips -z 512 512   "$SRC_ICON" --out "$ICONSET/icon_512x512.png"    >/dev/null
sips -z 1024 1024 "$SRC_ICON" --out "$ICONSET/icon_512x512@2x.png" >/dev/null
iconutil -c icns "$ICONSET" -o "$APP/Contents/Resources/AppIcon.icns"

cat > "$APP/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleDevelopmentRegion</key>
	<string>zh_CN</string>
	<key>CFBundleExecutable</key>
	<string>$BIN_NAME</string>
	<key>CFBundleIconFile</key>
	<string>AppIcon</string>
	<key>CFBundleIdentifier</key>
	<string>com.gostc.gui</string>
	<key>CFBundleInfoDictionaryVersion</key>
	<string>6.0</string>
	<key>CFBundleName</key>
	<string>$APP_NAME</string>
	<key>CFBundleDisplayName</key>
	<string>GOSTC</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleShortVersionString</key>
	<string>$VERSION</string>
	<key>CFBundleVersion</key>
	<string>$VERSION</string>
	<key>LSMinimumSystemVersion</key>
	<string>11.0</string>
	<key>LSUIElement</key>
	<true/>
	<key>NSHighResolutionCapable</key>
	<true/>
	<key>NSAppTransportSecurity</key>
	<dict>
		<key>NSAllowsLocalNetworking</key>
		<true/>
	</dict>
</dict>
</plist>
PLIST

echo "==> ad-hoc 签名（无开发者证书；arm64 可执行文件必须至少 ad-hoc 签名）"
codesign --force --deep --sign - "$APP"

echo "==> 制作 .dmg（含 Applications 快捷方式，拖拽安装）"
STAGE="$TMP_DIR/dmg"
mkdir -p "$STAGE"
cp -R "$APP" "$STAGE/"
ln -s /Applications "$STAGE/Applications"
DMG="$OUT_DIR/${APP_NAME}_${VERSION}_macos_universal.dmg"
hdiutil create -volname "$APP_NAME" -srcfolder "$STAGE" -ov -format UDZO "$DMG" >/dev/null

echo "完成：$DMG"
