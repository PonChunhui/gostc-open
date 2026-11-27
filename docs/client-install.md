# 客户端安装

客户端运行在需要被穿透的内网机器上，支持 Windows 图形版、Windows/Linux 命令行版。**安装包请从平台首页的「客户端下载」页面获取**，以保证与服务端版本匹配。

| 安装包 | 适用 |
| --- | --- |
| `gostc-gui.exe` | Windows 图形版（推荐新手） |
| `gostc-cli.exe` | Windows 命令行版 |
| `gostc-linux-amd64` | Linux 命令行版 |

## Windows 图形版（推荐新手）

1. 下载 `gostc-gui.exe`，双击运行（程序需要管理员权限，会弹出 UAC 确认）。
2. 首次运行如提示缺少 WebView2 运行时，按提示安装（Windows 11 一般已内置）。
3. 图形版按用途分为两类页面，注意区分角色：

### 客户端页（被访问端）

运行在**需要被访问的内网机器**上。点击「添加客户端」，填写：

- 名称、服务端地址（如 `gostc.shop`）、连接密钥（网页「客户端 → 连接密钥」获取）。
- 按需开启 TLS、开机自启。

启动后状态显示在线即连接成功。之后在网页上为该客户端创建隧道/转发，配置会实时下发。

### 隧道访客 / P2P 访客页（访问端）

运行在**发起访问的机器**上，用于访问别人（或自己另一台机器）创建的私有隧道 / P2P 隧道。点击「添加」，填写：

- 访问密钥（对方创建隧道后获得的密钥）、本地监听端口、服务端地址，按需开启 TLS。

启动后，访问本机监听端口即相当于访问隧道指向的内网服务。

> 图形版的数据保存在程序所在目录的 `data/` 下（client.json、tunnel.json、p2p.json 等），迁移时连同目录一起复制即可。

## 命令行版

### 基本连接（被访问端）

```shell
# Windows
gostc-cli.exe -addr gostc.shop -key 连接密钥

# Linux
./gostc -addr gostc.shop -key 连接密钥
```

常用参数：

| 参数 | 说明 | 默认 |
| --- | --- | --- |
| `-addr` | 服务端地址，如 `gostc.shop` 或 `127.0.0.1:8080` | gostc.shop |
| `-key` | 客户端连接密钥（网页「客户端 → 连接密钥」获取） | 空 |
| `--tls` | 服务端启用 SSL 时为 true，未启用设为 `--tls=false` | true |
| `-s` | 节点模式（公网机器使用，配合节点密钥） | false |
| `-v` | 隧道访客模式（访问端，配合 `-vts`） | false |
| `-p2p` | P2P 访客模式（访问端，配合 `-vts`） | false |
| `-vts` | 访客隧道列表，格式 `密钥:本地端口,密钥:本地端口` | 空 |
| `-console` | 日志输出到控制台 | false |
| `-version` | 查看版本 | - |

也支持环境变量 `GOSTC_CLIENT_ADDR`、`GOSTC_CLIENT_TLS` 覆盖地址与 TLS 设置。

### 访问端（隧道访客 / P2P 访客）

```shell
# 私有隧道访客：把密钥 aaaaaa 的隧道映射到本机 8080 端口
./gostc -addr gostc.shop -v -vts aaaaaa:8080

# P2P 访客
./gostc -addr gostc.shop -p2p -vts aaaaaa:8080
```

多条隧道用英文逗号分隔：`-vts aaaaaa:8080,bbbbbb:8081`。

### 注册为系统服务（Linux）

```shell
# 客户端
gostc install -addr gostc.shop -key ******
# 节点
gostc install -addr gostc.shop -s -key ******

systemctl start gostc     # 启动
systemctl stop gostc      # 停止
systemctl restart gostc   # 重启
systemctl status gostc    # 查看状态
```

更换密钥需先卸载再重新注册：

```shell
gostc uninstall
gostc install -addr gostc.shop -key 新密钥
```

> TLS 默认开启（`--tls` 默认 true），平台 `gostc.shop` 走 HTTPS，无需额外参数。服务名固定为 gostc，无法用系统服务同时运行多个客户端/节点。需要多开时请用 pm2、supervisor 等进程管理工具直接启动进程。

### Linux 安装为服务（从平台下载）

```shell
# 1. 从平台「客户端下载」获取 gostc-linux-amd64，上传到服务器后：
chmod +x gostc-linux-amd64
mv gostc-linux-amd64 /usr/local/bin/gostc

# 2. 注册为服务并启动
gostc install -addr gostc.shop -key 连接密钥
systemctl start gostc
```

## Docker 部署客户端

```yaml
services:
  client1:
    image: adonis996/gostc:latest
    restart: always
    network_mode: host
    container_name: gostc
    command:
      - --tls=true
      - -addr
      - gostc.shop
      - -key
      - "******" # 客户端密钥
```

节点部署在 command 中追加 `-s` 并使用节点密钥。

## 连接密钥在哪里

登录平台 → 「客户端」页面 → 对应客户端的「连接密钥」按钮。密钥与客户端一一对应，泄露后他人可冒用该客户端，请注意保管。
