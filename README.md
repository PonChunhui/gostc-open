# GOSTC

[中文](README.zh-CN.md) | English

An intranet penetration management platform built on FRP. It supports multiple users and multiple nodes, rate limiting, centralized configuration, and web-based management with changes applied in real time.

**Current platform: <https://gostc.shop/>**

## Features

- Multi-user, multi-node management
- Five tunnel types: private tunnel (TCP/UDP), port forwarding, domain resolution (HTTP/HTTPS), proxy tunnel (SOCKS5), and P2P
- Node load monitoring: real-time client / tunnel / connection counts per node
- Per-node tunnel capacity limits to prevent users from crowding onto a single node
- Traffic quotas and rate limiting per tunnel / user group
- Centralized configuration, edited on the web and applied in real time

## Preview

![dashboard](images/dashboard.png)

![node](images/node.png)

![config](images/config.png)

## Documentation

Detailed guides: <https://gostc.shop/docs/>

## Installation

Two options: **one-line script install (Linux)** and **Docker Compose**. Packages and images are published from this repository: [GitHub Releases](https://github.com/PonChunhui/gostc-open/releases) / [Docker Hub](https://hub.docker.com/u/adonis996).

### Option 1: One-line Script Install (Linux)

The script detects the system architecture, downloads the latest release from GitHub Releases, and installs it. Requires `bash`, `curl`, `sudo`, and `tar`/`unzip`.

**Install the admin server:**

```shell
curl -fsSL https://raw.githubusercontent.com/PonChunhui/gostc-open/main/install.sh | bash -s -- server
```

- Installs to `/usr/local/gostc-admin/` and registers the systemd service `gostc-admin`
- Start it with `systemctl start gostc-admin`, then open `http://SERVER_IP:8080` (default account / password: `admin/admin`)
- Data directory: `/usr/local/gostc-admin/data/` (config, SQLite database, logs)

**Install the client / node:**

```shell
curl -fsSL https://raw.githubusercontent.com/PonChunhui/gostc-open/main/install.sh | bash -s -- gostc
```

The script only installs the binary to `/usr/local/bin/gostc`. Connect it to the server manually:

```shell
# client (run on the intranet machine to be accessed)
gostc -addr gostc.shop -key CONNECTION_KEY

# node (run on a public machine; -s enables node mode)
gostc -addr gostc.shop -s -key NODE_KEY

# self-hosted server without SSL
gostc --tls=false -addr SERVER_IP:8080 -key CONNECTION_KEY
```

To register it as a system service, see "Register as a Service" below.

You can also download [install.sh](install.sh) and run `bash install.sh server` or `bash install.sh gostc` locally.

### Option 2: Docker Compose

Images published by this repository (built and pushed by CI on each release):

| Image | Description | Architectures |
| --- | --- | --- |
| `adonis996/gostc-admin` | Admin server | amd64 / arm64 |
| `adonis996/gostc` | Client / node | 386 / amd64 / armv7 / arm64 / s390x |
| `adonis996/gostc-proxy` | Gateway service (custom domains, optional) | 386 / amd64 / armv7 / arm64 / s390x |

Tag policy: stable releases publish both `latest` and a version tag (e.g. `2.0.12`); pre-releases (rc/beta) publish only the version tag and never overwrite `latest`.

**1. Deploy the admin server**

Create a `docker-compose.yml` (a ready-to-use template is provided at the repo root: [docker-compose.yml](docker-compose.yml)):

```yaml
services:
  gostc-admin:
    image: adonis996/gostc-admin:latest
    container_name: gostc-admin
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
```

```shell
docker compose up -d
```

- Open `http://SERVER_IP:8080` (default account / password: `admin/admin`)
- `./data` persists the config and database: `config.yaml` (auto-generated on first run), `data.db` (SQLite), and logs
- Override the listen address with the `GOSTC_ADMIN_ADDR` environment variable (e.g. `0.0.0.0:9090`; adjust the port mapping accordingly)
- For production, put Nginx/Caddy in front to enable HTTPS; clients connecting to a self-hosted server without SSL need `--tls=false`

**2. Deploy the client / node**

Run on the intranet machine to be accessed (client) or on a public machine (node):

```yaml
services:
  gostc:
    image: adonis996/gostc:latest
    container_name: gostc
    restart: unless-stopped
    network_mode: host
    command: ["-addr", "gostc.shop", "-key", "CONNECTION_KEY"]
```

- Node mode: append `"-s"` to `command` and use the node key
- Self-hosted server without SSL: `command: ["--tls=false", "-addr", "SERVER_IP:8080", "-key", "CONNECTION_KEY"]`
- `network_mode: host` lets the container use the host network directly so it can reach intranet target services; recommended
- Get the connection key from the web UI ("Clients → Connection Key"); get the node key from node management in the admin panel

**3. (Optional) Deploy the gateway service**

`gostc-proxy` provides the custom-domain feature and runs on a node, occupying ports 80/443:

```yaml
services:
  gostc-proxy:
    image: adonis996/gostc-proxy:latest
    container_name: gostc-proxy
    restart: unless-stopped
    network_mode: host
    volumes:
      - ./proxy-data:/app/data
```

`data/config.yaml` is auto-generated on first run (defaults: HTTP 80 / HTTPS 443). See the [documentation](https://gostc.shop/docs/) for details.

## Directory Structure

```text
-- /
    -- server   // backend code
    -- web      // frontend code
    -- client   // node and client code
    -- proxy    // gateway service, mainly for custom domains
```

## Build the Frontend

```shell
cd web
npm i
npm run build
```

The output is in `web/dist`. Zip the `dist` directory and copy it to `server/web/dist.zip` (entry names must use forward slashes, e.g. `dist/index.html`).

## Build the Backend

```shell
cd server
go build -ldflags "-s -w" -o server main.go
```

## Build the Node / Client

Current platform:

```shell
cd client
go build -ldflags "-s -w" -o gostc ./cli/
```

Cross-platform builds (with goreleaser):

```shell
cd client
goreleaser release --snapshot --clean
```

Windows GUI client:

```shell
cd client
go build -ldflags="-H windowsgui -w -s" -o gostc-gui.exe ./gui/
```

## Run the Server

```shell
./server
```

Default port: `8080`. Default account / password: `admin/admin`.

Program directory: `/usr/local/gostc-admin/`
Data directory: `/usr/local/gostc-admin/data/`

Manage the server service:

```shell
systemctl start gostc-admin    # start
systemctl stop gostc-admin     # stop
systemctl restart gostc-admin  # restart
systemctl status gostc-admin   # status
```

Uninstall the server:

```shell
/usr/local/gostc-admin/server service uninstall
rm -rf /usr/local/gostc-admin/server
```

## Connect a Client / Node to the Server

```shell
# node
./gostc -addr gostc.shop -s -key ******

# client
./gostc -addr gostc.shop -key ******
```

For a self-hosted server without SSL, set `--tls=false` and point `-addr` to your own server, e.g.:

```shell
./gostc --tls=false -addr 127.0.0.1:8080 -s -key ******
```

- `--tls`: set according to whether the server uses SSL
- `-addr`: the server address
- `-key`: the connection key of the client or node
- `-s`: run in node mode

## Register as a Service

```shell
gostc install --tls=false -addr 127.0.0.1:8080 -key ******     # client
gostc install --tls=false -addr 127.0.0.1:8080 -s -key ******  # node
# after install, start it with: systemctl start gostc
```

Manage the service:

```shell
systemctl start gostc    # start
systemctl stop gostc     # stop
systemctl restart gostc  # restart
systemctl status gostc   # status
```

To change the key, uninstall the service first and register again:

```shell
gostc uninstall
gostc install --tls=false -addr 127.0.0.1:8080 -s -key ******
```

Uninstall the client / node binary:

```shell
rm -rf /usr/local/bin/gostc
```

**Note: because the service name is fixed, you cannot run multiple clients/nodes as services at the same time. To run several instances, use a process manager such as pm2 or supervisor.**
