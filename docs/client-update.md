# 客户端自更新

客户端（命令行 / WebUI / 桌面 GUI 三端）支持从面板检查并升级新版本，共用同一套更新核心 `client/internal/updater`。

## 更新模式

设置项 `mode` 三档：

| 模式 | 行为 |
| --- | --- |
| `off` | 关闭自更新，不检查 |
| `notify`（默认） | 后台定时检查，发现新版本仅提示（托盘菜单 / WebUI 徽标），由用户确认后升级 |
| `auto` | 后台定时检查，发现新版本自动下载、校验、替换并重启 |

其它设置：`channel`（`stable`/`beta` 更新通道）、`intervalHours`（检查间隔，默认 24）、`panelUrl`（面板基址，留空用连接地址）、`autoRestart`（升级后是否重启，默认开）、`publicKey`（ed25519 签名公钥 hex，配置后强制校验签名）。

设置持久化在 `<BasePath>/data/update.json`，三端共享。

## 升级流程与安全

1. `GET /api/v1/public/client/version/latest?platform=&arch=&channel=&uid=&current=` 查询新版本（服务端按平台/架构/通道/灰度返回，`current` 防降级）。
2. 语义化版本比较（`v` 前缀、`-rcN` 预发布），低于或等于当前版本不升级。
3. 下载归档（zip / tar.gz，上限 512MB）。
4. **强制 SHA256 校验**（`sha256` 字段指归档哈希）；配置 `publicKey` 时额外校验 ed25519 签名。
5. 解出二进制（`gostc` / `gostc.exe`），原子替换：Windows 将运行中的 exe 改名 `.old` 再移入新文件；Unix 备份后 rename 覆盖。失败自动回滚。
6. 重启：交互模式 relaunch 自身；服务模式经服务管理器（kardianos）restart；`autoRestart=false` 时仅替换不重启。

## 命令行（CLI）

```bash
# 检查新版本（默认面板 gostc.shop）
gostc update -addr 127.0.0.1:8080 -tls=false

# 发现新版本立即升级（服务模式自动 restart）
gostc update -addr gostc.shop --yes

# 覆盖更新设置（持久化）
gostc update -mode auto -channel beta -interval 6 -url https://panel.example.com
```

运行旗标（`gostc -addr ... -key ...`）也支持 `-update-mode/-update-channel/-update-interval/-update-url`，并可用环境变量 `GOSTC_UPDATE_MODE/CHANNEL/INTERVAL/URL` 覆盖。client/server/ui 模式启动时自动开启后台检查。

## WebUI

Header 右上角齿轮按钮打开「客户端更新」抽屉：设置模式/通道/间隔/面板地址/重启/公钥，「检查更新」「保存设置」，发现新版本时徽标点亮并可「立即升级」。后端接口 `POST /api/config/update/{query,save,check,apply}`。

## 桌面 GUI

托盘菜单新增「检查更新」「升级到新版本」（发现新版本时显示）。`notify` 模式后台检查到新版本会点亮托盘升级菜单与提示。

## 服务端发版规范

在后台「客户端版本管理」发布版本：

- `version`：语义化版本（如 `v2.1.0`）。
- `platform`/`arch`：目标平台架构（`windows`/`linux`/`darwin`，`amd64`/`arm64`）。
- `downloadUrl`：归档下载地址（goreleaser 产物 `gostc_<target>.zip`/`.tar.gz`）。
- `sha256`：**归档文件**的 SHA256（必填，客户端强制校验）。
- `signature`：可选，归档的 ed25519 签名 hex（客户端配置公钥后校验）。
- `channel`：`stable`/`beta`；`grayPercent`：灰度百分比（按 uid 哈希分桶）。
- `enable`：启用后才会被 `latest` 返回。

> Android 端禁用自更新（由应用商店分发）。旧客户端不受影响：未升级的二进制忽略新字段，服务端不强制升级。
