# GOSTC

中文 | [English](README.md)

基于 FRP 开发的内网穿透管理平台，支持多用户、多节点，支持速率限制，中心化配置，通过网页修改配置，实时生效。

**当前平台地址：<https://gostc.shop/>**

## 功能特性

- 多用户、多节点管理
- 五种隧道类型：私有隧道（TCP/UDP）、端口转发、域名解析（HTTP/HTTPS）、代理隧道（SOCKS5）、P2P
- 节点负载监控：实时展示每个节点的客户端数、运行隧道数、连接数
- 节点隧道数量上限，避免大量用户挤占同一节点
- 按隧道/用户组设置流量配额与速率限制
- 中心化配置，网页修改，实时生效

## 预览

![dashboard](images/dashboard.png)

![node](images/node.png)

![config](images/config.png)

## 文档

详细教程：<https://gostc.shop/docs/>

## 目录结构

```text
-- /
    -- server   // 后端项目代码
    -- web      // 前端项目代码
    -- client   // 节点和客户端项目代码
    -- proxy    // 网关服务，主要实现自定义域名功能
```

## 编译前端代码

```shell
cd web
npm i
npm run build
```

编译后的文件在 `web/dist` 目录中，压缩 dist 目录，将 dist.zip 复制到 `server/web/dist.zip`（压缩包内条目路径需使用正斜杠，如 `dist/index.html`）。

## 编译后端代码

```shell
cd server
go build -ldflags "-s -w" -o server main.go
```

## 编译节点/客户端

当前平台：

```shell
cd client
go build -ldflags "-s -w" -o gostc ./cli/
```

多平台编译（借助 goreleaser 工具）：

```shell
cd client
goreleaser release --snapshot --clean
```

Windows GUI 客户端：

```shell
cd client
go build -ldflags="-H windowsgui -w -s" -o gostc-gui.exe ./gui/
```

## 运行后台管理（服务端）

```shell
./server
```

默认端口：`8080`，默认账号密码：`admin/admin`。

程序目录：`/usr/local/gostc-admin/`
数据目录：`/usr/local/gostc-admin/data/`

管理服务：

```shell
systemctl start gostc-admin    # 启动
systemctl stop gostc-admin     # 停止
systemctl restart gostc-admin  # 重启
systemctl status gostc-admin   # 查看状态
```

服务端卸载：

```shell
/usr/local/gostc-admin/server service uninstall
rm -rf /usr/local/gostc-admin/server
```

## 客户端和节点如何指定服务端地址

```shell
# 节点
./gostc -addr gostc.shop -s -key ******

# 客户端
./gostc -addr gostc.shop -key ******
```

如果自建的后台管理服务不启用 SSL，需要设置 `--tls=false`，并将 `-addr` 指向实际的服务端地址，例如：

```shell
./gostc --tls=false -addr 127.0.0.1:8080 -s -key ******
```

- `--tls`：根据服务端是否使用 SSL 设置
- `-addr`：服务端访问地址
- `-key`：客户端或节点的连接密钥
- `-s`：以节点模式运行

## 将客户端/节点注册为服务

```shell
gostc install --tls=false -addr 127.0.0.1:8080 -key ******     # 客户端
gostc install --tls=false -addr 127.0.0.1:8080 -s -key ******  # 节点
# install 后，需要 systemctl start gostc 启动服务
```

管理服务：

```shell
systemctl start gostc    # 启动
systemctl stop gostc     # 停止
systemctl restart gostc  # 重启
systemctl status gostc   # 查看状态
```

如需更换密钥，需要先卸载服务，然后重新注册：

```shell
gostc uninstall
gostc install --tls=false -addr 127.0.0.1:8080 -s -key ******
```

客户端/节点卸载：

```shell
rm -rf /usr/local/bin/gostc
```

**注意：由于服务名称重复，无法同时运行多个客户端/节点服务，如需启动多个实例，请通过 pm2、supervisor 等进程管理工具启动。**
