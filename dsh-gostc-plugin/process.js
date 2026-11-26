/**
 * Host-side lifecycle manager for the gostc command-line client.
 *
 * Responsibilities:
 *  - provision the CLI binary (download from the user's own gostc platform),
 *    selecting the package that matches the host OS + CPU architecture,
 *  - reuse an already-healthy backend or launch a plugin-managed child process,
 *  - expose start/stop/restart and Windows-service install/uninstall (UAC),
 *  - terminate the managed child when the plugin is disposed.
 *
 * Uses ctx.subprocess (managed argv spawn, bounded output, idempotent terminate).
 * The ordinary SubprocessHandle exposes no pid, so status reports mode/phase only.
 */
import fs from 'node:fs';
import fsp from 'node:fs/promises';
import path from 'node:path';
import os from 'node:os';

const MAX_LOG = 8000;
const MAX_RESTARTS = 3;

function defaultDataDir() {
  const local = process.env.LOCALAPPDATA;
  if (local) return path.join(local, 'dsh-gostc');
  return path.join(os.homedir(), '.dsh-gostc');
}

/** Detect the host OS + CPU architecture (Windows WoW64 aware). */
export function hostTarget() {
  const osName = process.platform; // win32|linux|darwin|freebsd|...
  let arch = process.arch; // x64|arm64|ia32|arm|...
  if (osName === 'win32') {
    const pa = String(process.env.PROCESSOR_ARCHITEW6432 || process.env.PROCESSOR_ARCHITECTURE || '').toUpperCase();
    if (pa === 'AMD64') arch = 'x64';
    else if (pa === 'ARM64') arch = 'arm64';
    else if (pa === 'x86') arch = process.env.PROCESSOR_ARCHITEW6432 ? 'x64' : 'ia32';
  }
  return { os: osName, arch };
}

const SUPPORTED_OS = new Set(['win32', 'linux', 'darwin', 'freebsd']);

/** Canonical OS inferred from a package filename, or null when unspecified. */
export function osOfName(lower) {
  if (/windows|(^|[-_.])win([-_.]|$)/.test(lower)) return 'win32';
  if (/linux/.test(lower)) return 'linux';
  if (/darwin|macos|osx|(^|[-_.])mac([-_.]|$)/.test(lower)) return 'darwin';
  if (/freebsd/.test(lower)) return 'freebsd';
  return null;
}

/** Canonical CPU arch inferred from a package filename, or null when unspecified. */
export function archOfName(lower) {
  if (/arm64|aarch64/.test(lower)) return 'arm64';
  if (/amd64|x86_64|x64/.test(lower)) return 'x64';
  if (/armv[567]|armhf|arm_?[567]|(^|[-_.])arm([-_.]|$)/.test(lower)) return 'arm';
  if (/386|i686|(^|[-_.])x86([-_.]|$)/.test(lower)) return 'ia32';
  return null;
}

const ARCHIVE_RE = /\.(zip|tar\.gz|tgz|gz|tar|xz|7z|rar|bz2)$/i;

/**
 * Pick the download entry matching the host OS + arch. Prefers the CLI build and
 * a raw binary over an archive; rejects entries that explicitly name a different
 * OS or arch. Windows names without an OS token (e.g. gostc-cli.exe) are accepted
 * implicitly via the .exe extension.
 */
export function selectPackage(list, target) {
  let best = null;
  let bestScore = -Infinity;
  for (const it of list) {
    if (!it || typeof it.name !== 'string') continue;
    const lower = it.name.toLowerCase();
    const tokens = lower.split(/[^a-z0-9]+/).filter(Boolean);
    const hasExe = /\.exe$/.test(lower);
    const isArchive = ARCHIVE_RE.test(lower);
    const nameOs = osOfName(lower);
    const nameArch = archOfName(lower);
    if (nameOs && nameOs !== target.os) continue; // explicit OS mismatch
    if (nameArch && nameArch !== target.arch) continue; // explicit arch mismatch
    const osMatch = nameOs === target.os || (!nameOs && target.os === 'win32' && hasExe);
    if (!osMatch) continue;
    let score = 0;
    if (nameArch === target.arch) score += 4; // explicit arch match beats unspecified
    if (tokens.includes('cli')) score += 3;
    if (tokens.includes('gui')) score -= 6; // we launch the headless CLI, not the tray GUI
    if (!isArchive) score += 2; else score -= 4; // raw binary preferred; archives need extraction
    if (target.os === 'win32' && hasExe) score += 1;
    if (score > bestScore) { bestScore = score; best = it; }
  }
  return best ? best.name : null;
}

/** Validate the downloaded bytes against the host executable magic. */
export function checkMagic(buf, osName) {
  if (buf.length < 1024 * 1024) return `下载内容过小（${buf.length} 字节），疑似错误响应`;
  const b0 = buf[0]; const b1 = buf[1]; const b2 = buf[2]; const b3 = buf[3];
  if (b0 === 0x50 && b1 === 0x4b) return '下载到的是压缩包（zip），请提供解压后的二进制安装包';
  if (osName === 'win32') {
    if (b0 === 0x4d && b1 === 0x5a) return null; // MZ
    return '下载内容不是有效的 Windows 可执行文件（缺少 MZ 头）';
  }
  if (osName === 'linux' || osName === 'freebsd') {
    if (b0 === 0x7f && b1 === 0x45 && b2 === 0x4c && b3 === 0x46) return null; // ELF
    return '下载内容不是有效的 ELF 可执行文件';
  }
  if (osName === 'darwin') {
    const magic = buf.readUInt32BE(0);
    const ok = magic === 0xfeedface || magic === 0xfeedfacf || magic === 0xcafebabe || magic === 0xcffaedfe || magic === 0xcefaedfe;
    if (ok) return null; // Mach-O / fat
    return '下载内容不是有效的 macOS 可执行文件（Mach-O）';
  }
  return null;
}

export function createProcessManager(ctx, config) {
  const backend = String(config?.backend || 'http://127.0.0.1:18080').replace(/\/+$/, '');
  const platformUrl = String(config?.platformUrl || '').replace(/\/+$/, '');
  const clientPackage = String(config?.clientPackage || '');
  const webAddr = String(config?.webAddr || '127.0.0.1:18080');
  const webUser = String(config?.webUser || '');
  const webPwd = String(config?.webPwd || '');
  const autoStart = config?.autoStart !== false;
  const username = String(config?.username || '');
  const password = String(config?.password || '');
  const authHeader = username
    ? 'Basic ' + Buffer.from(`${username}:${password}`).toString('base64')
    : undefined;
  const dataDir = String(config?.dataDir || '') || defaultDataDir();
  const target = hostTarget();
  const isWindows = target.os === 'win32';
  const supported = SUPPORTED_OS.has(target.os);
  const binName = isWindows ? 'gostc-cli.exe' : 'gostc-cli';
  const exePath = path.join(dataDir, binName);

  const state = {
    phase: 'idle', // idle|downloading|starting|running|stopped|error|unsupported
    mode: null, // managed|external|service
    version: '',
    error: null,
    downloaded: false,
    retries: 0,
    supported,
    target, // { os, arch }
    platformUrl,
    clientPackage,
    dataDir,
    exePath,
    webAddr,
    backend,
    autoStart,
  };

  let handle = null;
  let intentionalStop = false;
  let disposed = false;
  let restartTimer = null;
  let busy = false;

  const getSubprocess = () => ctx.get('subprocess');

  function readLogTail() {
    try {
      const reader = handle && handle.collected && handle.collected.stdout;
      if (reader) return reader.readFrom(0).text.slice(-MAX_LOG);
    } catch {}
    return '';
  }

  function snapshot() {
    return Object.assign({}, state, {
      hasExe: fs.existsSync(exePath),
      logTail: readLogTail(),
      busy,
    });
  }

  function fail(message) {
    state.phase = 'error';
    state.error = String(message);
  }

  async function probe() {
    const ctrl = new AbortController();
    const timer = setTimeout(() => ctrl.abort(), 2500);
    try {
      const res = await fetch(`${backend}/api/config/query`, {
        method: 'POST',
        headers: {
          'content-type': 'application/json',
          accept: 'application/json',
          ...(authHeader ? { authorization: authHeader } : {}),
        },
        body: '{}',
        signal: ctrl.signal,
      });
      const j = await res.json();
      if (j && j.code === 0) {
        return { healthy: true, version: (j.data && j.data.version) || '' };
      }
      return { healthy: false, version: '' };
    } catch {
      return { healthy: false, version: '' };
    } finally {
      clearTimeout(timer);
    }
  }

  async function runVersion() {
    const sp = getSubprocess();
    if (!sp) throw new Error('subprocess 服务不可用');
    const h = sp.spawn({
      argv: [exePath, '-version'],
      cwd: dataDir,
      stdio: { stdin: 'ignore', stdout: { maxBytes: 16 * 1024 }, stderr: { maxBytes: 16 * 1024 } },
      graceMs: 4000,
    });
    const { exitCode } = await h.done;
    const out = h.collected.stdout ? h.collected.stdout.readFrom(0).text : '';
    if (exitCode !== 0) throw new Error(`客户端 -version 退出码 ${exitCode}: ${out.slice(0, 200)}`);
    return out.trim().split(/\r?\n/)[0].trim();
  }

  async function discoverPackage() {
    const res = await fetch(`${platformUrl}/api/v1/public/download/list`, {
      method: 'POST',
      headers: { 'content-type': 'application/json', accept: 'application/json' },
      body: '{}',
    });
    const j = await res.json();
    if (!j || j.code !== 0 || !Array.isArray(j.data)) {
      throw new Error('平台下载列表获取失败：' + ((j && j.msg) || res.status));
    }
    const name = selectPackage(j.data, target);
    if (!name) {
      throw new Error(`平台 downloads 中未找到匹配当前系统（${target.os}/${target.arch}）的客户端安装包`);
    }
    return name;
  }

  async function download(name) {
    const url = `${platformUrl}/api/v1/public/download/file?name=${encodeURIComponent(name)}`;
    const res = await fetch(url, { headers: { accept: 'application/octet-stream' } });
    if (!res.ok) throw new Error(`下载失败 HTTP ${res.status}（${name}）`);
    const buf = Buffer.from(await res.arrayBuffer());
    const magicErr = checkMagic(buf, target.os);
    if (magicErr) throw new Error(magicErr);
    await fsp.mkdir(dataDir, { recursive: true });
    const tmp = path.join(dataDir, `${binName}.tmp-${process.pid}`);
    await fsp.writeFile(tmp, buf);
    await fsp.rename(tmp, exePath);
    if (!isWindows) { try { await fsp.chmod(exePath, 0o755); } catch {} }
    state.downloaded = true;
  }

  async function ensureClient(force) {
    if (!supported) { fail(`当前系统（${target.os}/${target.arch}）暂不支持自动安装客户端`); return { ok: false }; }
    try {
      await fsp.mkdir(dataDir, { recursive: true });
      if (fs.existsSync(exePath) && !force) {
        try {
          state.version = await runVersion();
          state.error = null;
          return { ok: true, downloaded: false };
        } catch {
          // existing binary unusable → fall through to re-download
        }
      }
      if (!platformUrl) {
        if (fs.existsSync(exePath)) {
          state.error = null;
          return { ok: true, downloaded: false };
        }
        fail('未配置平台地址 platformUrl，无法下载客户端');
        return { ok: false };
      }
      state.phase = 'downloading';
      state.error = null;
      const name = clientPackage || (await discoverPackage());
      state.clientPackage = name;
      await download(name);
      state.version = await runVersion();
      return { ok: true, downloaded: true };
    } catch (err) {
      fail(err.message || String(err));
      return { ok: false };
    }
  }

  function watchExit(h) {
    h.done.then((outcome) => {
      if (disposed || intentionalStop) return;
      if (handle !== h) return;
      // unexpected exit → bounded auto-restart
      state.retries += 1;
      state.mode = null;
      if (state.retries > MAX_RESTARTS) {
        fail(`客户端反复退出（exitCode=${outcome.exitCode}），已停止自动重启`);
        state.phase = 'error';
        return;
      }
      state.phase = 'stopped';
      state.error = `客户端退出（exitCode=${outcome.exitCode}），${state.retries}/${MAX_RESTARTS} 次自动重启`;
      const delay = 1000 * state.retries;
      restartTimer = setTimeout(() => { restartTimer = null; start().catch(() => {}); }, delay);
    }).catch(() => {});
  }

  async function waitReady(timeoutMs) {
    const deadline = Date.now() + timeoutMs;
    while (Date.now() < deadline && !disposed) {
      const p = await probe();
      if (p.healthy) return p;
      await new Promise((r) => setTimeout(r, 500));
    }
    return { healthy: false, version: '' };
  }

  async function start() {
    if (busy) return snapshot();
    if (!supported) { state.phase = 'unsupported'; return snapshot(); }
    busy = true;
    intentionalStop = false;
    try {
      // 1) reuse an already-healthy backend (e.g. a manually started gostc-gui.exe)
      const existing = await probe();
      if (existing.healthy) {
        state.phase = 'running';
        state.mode = 'external';
        state.version = existing.version;
        state.error = null;
        state.retries = 0;
        return snapshot();
      }
      // 2) provision the binary
      const ensured = await ensureClient(false);
      if (!ensured.ok) return snapshot();
      // 3) launch a managed child
      const sp = getSubprocess();
      if (!sp) { fail('subprocess 服务不可用，无法拉起客户端'); return snapshot(); }
      state.phase = 'starting';
      state.error = null;
      const argv = [exePath, '-web-addr', webAddr, '-console'];
      if (webUser) argv.push('-web-user', webUser, '-web-pwd', webPwd);
      handle = sp.spawn({
        argv,
        cwd: dataDir,
        stdio: { stdin: 'ignore', stdout: { maxBytes: 256 * 1024 }, stderr: { maxBytes: 256 * 1024 } },
        graceMs: 4000,
      });
      watchExit(handle);
      const ready = await waitReady(15000);
      if (disposed) return snapshot();
      if (ready.healthy) {
        state.phase = 'running';
        state.mode = 'managed';
        state.version = ready.version || state.version;
        state.error = null;
        state.retries = 0;
      } else {
        fail('客户端启动后 15 秒内未就绪，请查看日志');
      }
      return snapshot();
    } catch (err) {
      fail(err.message || String(err));
      return snapshot();
    } finally {
      busy = false;
    }
  }

  async function stop() {
    intentionalStop = true;
    if (restartTimer) { clearTimeout(restartTimer); restartTimer = null; }
    const h = handle;
    handle = null;
    if (h) {
      try { h.terminate(); } catch {}
      try { await h.waitForExit(); } catch {}
    }
    state.retries = 0;
    if (state.mode === 'managed' || state.phase === 'running' || state.phase === 'starting') {
      state.phase = 'stopped';
      state.mode = null;
    }
    state.error = null;
    return snapshot();
  }

  async function restart() {
    await stop();
    intentionalStop = false;
    return start();
  }

  function psQuote(s) {
    return "'" + String(s).replace(/'/g, "''") + "'";
  }

  async function elevate(args) {
    const sp = getSubprocess();
    if (!sp) throw new Error('subprocess 服务不可用');
    const argList = '@(' + args.map(psQuote).join(',') + ')';
    const script = `Start-Process -FilePath ${psQuote(exePath)} -ArgumentList ${argList} -Verb RunAs -WindowStyle Hidden -Wait`;
    const h = sp.spawn({
      argv: ['powershell.exe', '-NoProfile', '-Command', script],
      cwd: dataDir,
      stdio: { stdin: 'ignore', stdout: { maxBytes: 16 * 1024 }, stderr: { maxBytes: 16 * 1024 } },
      graceMs: 60000,
    });
    const { exitCode } = await h.done;
    const errText = h.collected.stderr ? h.collected.stderr.readFrom(0).text : '';
    if (exitCode !== 0) {
      throw new Error(`提权操作失败（退出码 ${exitCode}，可能取消了 UAC 授权）：${errText.slice(0, 200)}`);
    }
  }

  async function serviceInstall() {
    if (!isWindows) { fail('安装为系统服务当前仅支持 Windows（UAC 提权）'); return snapshot(); }
    try {
      const ensured = await ensureClient(false);
      if (!ensured.ok) return snapshot();
      state.phase = 'starting';
      state.error = null;
      const args = ['install', '-web-addr', webAddr];
      if (webUser) args.push('-web-user', webUser, '-web-pwd', webPwd);
      await elevate(args);
      // service may take a moment to start
      const ready = await waitReady(10000);
      if (ready.healthy) {
        state.phase = 'running';
        state.mode = 'service';
        state.version = ready.version || state.version;
        state.error = null;
      } else {
        state.error = '服务已安装，但后端尚未就绪，可稍后刷新或手动启动服务';
        state.phase = 'stopped';
      }
      return snapshot();
    } catch (err) {
      fail(err.message || String(err));
      return snapshot();
    }
  }

  async function serviceUninstall() {
    if (!isWindows) { fail('卸载系统服务当前仅支持 Windows（UAC 提权）'); return snapshot(); }
    try {
      state.error = null;
      await elevate(['uninstall']);
      if (state.mode === 'service') { state.mode = null; state.phase = 'stopped'; }
      state.error = null;
      return snapshot();
    } catch (err) {
      fail(err.message || String(err));
      return snapshot();
    }
  }

  async function kickoff() {
    if (!supported) { state.phase = 'unsupported'; return; }
    // wait briefly for the subprocess service to appear during boot
    for (let i = 0; i < 20 && !getSubprocess() && !disposed; i += 1) {
      await new Promise((r) => setTimeout(r, 250));
    }
    if (disposed) return;
    // reuse an already-healthy backend first (no provisioning needed)
    const existing = await probe();
    if (disposed) return;
    if (existing.healthy) {
      state.phase = 'running';
      state.mode = 'external';
      state.version = existing.version;
      state.error = null;
      state.retries = 0;
      return;
    }
    // otherwise auto-install the client, then launch when autoStart is on
    const ensured = await ensureClient(false);
    if (disposed) return;
    if (ensured.ok && autoStart) await start();
  }

  function dispose() {
    disposed = true;
    intentionalStop = true;
    if (restartTimer) { clearTimeout(restartTimer); restartTimer = null; }
    const h = handle;
    handle = null;
    if (h) { try { h.terminate(); } catch {} }
  }

  return {
    snapshot,
    probe,
    start,
    stop,
    restart,
    ensureClient,
    serviceInstall,
    serviceUninstall,
    kickoff,
    dispose,
  };
}