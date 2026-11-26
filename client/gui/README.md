# GOSTC GUI 客户端构建说明

图形界面（托盘 + 内嵌 Web 管理窗口）按平台拆分实现：

| 平台 | 内嵌窗口 | 托盘 | 开机自启 | 设置窗口 |
| --- | --- | --- | --- | --- |
| Windows | WebView2（`window/webview_windows.go`） | energye/systray | 注册表 Run 项 | gioui 窗口 |
| macOS | WKWebView（`window/webview_darwin.go/.m`，自研 cgo 封装，无新增 Go 依赖） | energye/systray | LaunchAgent | 原生 NSAlert 对话框 |
| 其他 | 系统浏览器打开管理界面 | energye/systray | 暂不支持 | 暂不支持 |

前端资源已内嵌在 `client/webui/backend/web/dist.zip`，GUI 构建**不需要 Node.js**。

## Windows

```bat
build.bat
```

产物 `gostc-gui.exe`（需要系统已安装 WebView2 Runtime，Win10/11 一般自带）。

## macOS

GUI 依赖 cgo（Cocoa / WebKit），**必须在 macOS 上编译**，无法从 Windows/Linux 直接交叉出 .app。

### 本机打包（推荐）

环境要求：macOS 11+、Xcode Command Line Tools（`xcode-select --install`）、Go ≥ 1.24。

```bash
cd client
bash gui/packaging/macos/package.sh 2.0.11
```

脚本会：

1. 编译 arm64 + 交叉编译 amd64（`clang -arch x86_64`），`lipo` 合并为 universal 二进制；
2. 用 `sips`/`iconutil` 从 `gui/static/logo.png` 生成 .icns，组装 `GOSTC.app`；
3. ad-hoc 签名（`codesign --sign -`）；
4. 产出 `client/gui/dist/GOSTC_<版本>_macos_universal.dmg`（含 Applications 拖拽安装快捷方式）。

只想要可执行文件时：`cd client && CGO_ENABLED=1 go build -o gostc-gui ./gui/`。

### CI 自动打包

- `.github/workflows/gui-macos.yml`：推送 `v*` tag（或手动 workflow_dispatch）时在 `macos-14`
  runner 上执行同一脚本，dmg 作为 artifact 上传，并附加到对应 GitHub Release。
  注意：部分 GitHub 账号会被平台风控限制 macOS runner（任务秒取消）；此时可用下面的 Cirrus CI。
- `.cirrus.yml`（根目录）：公开仓库在 [cirrus-ci.com](https://cirrus-ci.com) 免费使用 macOS runner。
  用 GitHub 登录并给本仓库安装 Cirrus CI App 后，推送即触发，dmg 在任务页面的 Artifacts 下载。

### 首次打开（Gatekeeper）

当前未配置 Apple 开发者证书，dmg 内应用仅 ad-hoc 签名、未公证。从浏览器下载后首次打开：

- 右键 `GOSTC.app` →「打开」→ 再点「打开」；或
- 终端执行：`xattr -dr com.apple.quarantine /Applications/GOSTC.app`

后续如有开发者证书，可在打包脚本中加入正式签名与 `xcrun notarytool` 公证流程。

### macOS 行为说明

- 数据目录：`~/Library/Application Support/GOSTC/data/`（config.yaml、client.json、tunnel.json、p2p.json 等）。
  .app 包内不可写且受 Gatekeeper 路径随机化影响，故不放在可执行文件旁。
- 关闭窗口 = 隐藏回托盘，程序继续运行；托盘菜单「退出」才真正退出。
- 「开机自启」写入 `~/Library/LaunchAgents/com.gostc.gui.plist`（登录时启动）。
- 托盘图标使用 `gui/static/logo.png`（macOS 对 ico 解码不可靠）。
