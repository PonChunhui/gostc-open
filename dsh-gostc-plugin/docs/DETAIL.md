# Gostc

A global management panel for the **gostc** intranet-penetration client
(`gostc-open`, Apache-2.0). It brings the gostc webui's management surface into
the DeepSeek Harness web client: client connections, private tunnels, P2P links
and runtime logs, with create / edit / delete / start / stop.

The plugin also **provisions and launches the gostc command-line client itself**:
it downloads the CLI installer from your own gostc platform, reuses an
already-running backend when one is healthy, and otherwise starts a
plugin-managed child process. A **Backend** tab exposes start / stop / restart,
client download / update, and Windows-service install / uninstall.

The actual tunneling is still performed by the gostc process; the plugin manages
it through the gostc webui HTTP API and controls the process lifecycle.

## Requirements

- Automatic provisioning / launch supports **Windows, Linux, macOS and FreeBSD**.
  The plugin detects the host **OS and CPU architecture** (Windows WoW64 aware)
  and downloads the matching installer. Installing as a **Windows service** (UAC)
  is Windows-only.
- A gostc **platform (server)** that exposes the public download API
  (`POST /api/v1/public/download/list`, `GET /api/v1/public/download/file?name=...`)
  with the CLI installer in its `downloads` directory. Set its base URL as
  `platformUrl`.
- If you already run a gostc backend (e.g. the tray app `gostc-gui.exe`), the
  plugin detects and reuses it; no download or launch happens.

## How it works

The Host half registers two same-origin prefix routes on the Harness web server:

- `/gostc/api/<kind>/<op>` forwards gostc webui API calls to the configured
  backend server-side. This avoids the gostc backend's lack of CORS support and
  needs no rebuild of gostc. The browser panel fetches the document-relative form
  (`gostc/api/...`), which also works under the Desktop `file://` fetch bridge.
- `/gostc/process/<op>` controls the client lifecycle (status / start / stop /
  restart / download / service install / uninstall).

Only known subpaths are forwarded (allowlist), and every request passes the
Harness connection trust fence. The managed child is spawned through
`ctx.subprocess` (bounded output, idempotent terminate) and is terminated when the
plugin is disposed.

### Provisioning flow

1. Probe `backend` (`/api/config/query`). If healthy → reuse it (`external` mode).
2. Otherwise ensure the CLI binary exists in `dataDir`; if missing (or on
   **Update client**), download it from `platformUrl`. The package is auto-picked
   to match the host **OS + CPU arch** (or forced via `clientPackage`), preferring
   the CLI build and a raw binary over an archive. Recognized naming includes
   `gostc-cli.exe`, `gostc-linux-amd64`, `gostc-darwin-arm64`, and goreleaser-style
   `gostc_windows_amd64` / `gostc_linux_arm64`. The bytes are sanity-checked against
   the platform executable magic (PE / ELF / Mach-O), written atomically, and made
   executable on non-Windows hosts.
3. Launch `gostc-cli.exe -web-addr <webAddr> -console` as a managed child and poll
   until the backend is ready (`managed` mode). Unexpected exits trigger bounded
   auto-restart.

**Install as service** runs `gostc-cli.exe install -web-addr ...` elevated (UAC)
so the backend runs at boot independently of the plugin (`service` mode).

## Configuration

Edit the plugin row's `config` in `cordis.patch.yml` (or the Plugins config page):

| Key | Default | Meaning |
| --- | --- | --- |
| `backend` | `http://127.0.0.1:18080` | Base URL of the gostc webui backend to probe / proxy. |
| `username` | `` | Optional HTTP Basic auth user for the backend (CLI `-web-user`). |
| `password` | `` | Optional HTTP Basic auth password for the backend (CLI `-web-pwd`). |
| `platformUrl` | `` | Your gostc platform base URL used to download the CLI installer. Empty disables auto-download. |
| `clientPackage` | `` | Exact installer filename in the platform `downloads` dir; empty auto-picks the Windows CLI package. |
| `dataDir` | `` | Install / data directory; empty defaults to `%LOCALAPPDATA%/dsh-gostc`. |
| `webAddr` | `127.0.0.1:18080` | webui listen address used when the plugin launches the client. |
| `webUser` | `` | Optional webui Basic auth user set on the launched client. |
| `webPwd` | `` | Optional webui Basic auth password set on the launched client. |
| `autoStart` | `true` | On host activation, auto-provision and launch (or reuse a healthy backend). |

## Notes

- Editing or deleting a running client/tunnel is refused by gostc; stop it first.
  The panel disables those actions while an entry is running.
- When the backend is unreachable the panel shows a friendly connection error
  with a retry action instead of crashing.
- Service install / uninstall triggers a Windows UAC prompt; cancelling it is
  reported as an error in the Backend tab.