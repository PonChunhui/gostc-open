/**
 * Host half of @local/dsh-gostc.
 *
 * Two prefix routes on the Harness web server (both same-origin, connection-fenced):
 *  - /gostc/api/<kind>/<op>  → forwards gostc webui API calls to a running backend
 *    server-side (avoids the gostc gin backend's lack of CORS; no gostc rebuild).
 *  - /gostc/process/<op>     → plugin lifecycle control (provision/start/stop the
 *    gostc command-line client, Windows-service install/uninstall).
 *
 * Pattern follows @deepseek-ai/dsh-host-open-in-app: raw webServer routes (not a
 * Typert Remote), the connection trust fence on every request, a bounded JSON
 * body, and fixed allowlists of subpaths (anti open-proxy / SSRF).
 *
 * NOTE: the webServer prefix matcher tests `pathname.startsWith(prefix + '/')`,
 * so a registered prefix must NOT carry a trailing slash. The *_PREFIX constants
 * keep their trailing slash for client routing and subpath slicing; REGISTER_*
 * strip it for registration.
 */
import {
  GOSTC_API_PREFIX, GOSTC_SUBPATHS,
  GOSTC_PROCESS_PREFIX, GOSTC_PROCESS_SUBPATHS,
} from './shared.js';
import { createProcessManager } from './process.js';
import { writeFileSync } from 'node:fs';

const DIAG = 'C:/code/gostc-open-main/dsh-gostc-plugin/';
const diag = (f, o) => { try { writeFileSync(DIAG + f, JSON.stringify(Object.assign({ ts: new Date().toISOString(), pid: process.pid }, o))); } catch {} };
diag('diag-1-module.json', { stage: 'module-loaded' });

export const name = 'gostc';

const ALLOWED = new Set(GOSTC_SUBPATHS);
const PROCESS_ALLOWED = new Set(GOSTC_PROCESS_SUBPATHS);
const MAX_BODY = 64 * 1024; // 64 KiB
const REGISTER_PATH = GOSTC_API_PREFIX.replace(/\/+$/, ''); // '/gostc/api'
const REGISTER_PROCESS_PATH = GOSTC_PROCESS_PREFIX.replace(/\/+$/, ''); // '/gostc/process'

function sendJson(res, status, payload) {
  const body = JSON.stringify(payload);
  res.statusCode = status;
  res.setHeader('content-type', 'application/json; charset=utf-8');
  res.setHeader('cache-control', 'no-store');
  res.end(body);
}

function sendMethodNotAllowed(res, allow) {
  res.statusCode = 405;
  res.setHeader('allow', allow);
  res.setHeader('content-type', 'application/json; charset=utf-8');
  res.end(JSON.stringify({ code: 1, msg: `method not allowed, use ${allow}`, data: null }));
}

/** Read a bounded request body; resolves null when it exceeds MAX_BODY (drained). */
function readBoundedBody(req) {
  return new Promise((resolve, reject) => {
    let size = 0;
    let overflow = false;
    const chunks = [];
    req.on('data', (chunk) => {
      size += chunk.length;
      if (size > MAX_BODY) {
        overflow = true;
        chunks.length = 0;
        req.resume();
        return;
      }
      if (!overflow) chunks.push(chunk);
    });
    req.on('end', () => resolve(overflow ? null : Buffer.concat(chunks).toString('utf8')));
    req.on('error', reject);
  });
}

export function apply(ctx, config) {
  diag('diag-2-apply.json', { stage: 'apply', hasWebServer: !!ctx.get('webServer') });
  const backend = String(config?.backend || 'http://127.0.0.1:18080').replace(/\/+$/, '');
  const username = String(config?.username || '');
  const password = String(config?.password || '');
  const authHeader = username
    ? 'Basic ' + Buffer.from(`${username}:${password}`).toString('base64')
    : undefined;

  // Lifecycle manager is created once per apply; kickoff + dispose are owned by
  // a dedicated effect so route (re)registration never duplicates the process.
  const manager = createProcessManager(ctx, config);
  ctx.effect(() => {
    manager.kickoff().catch(() => {});
    return () => manager.dispose();
  }, 'gostc-process-lifecycle');

  /** Register both prefix routes on whichever context exposes the web server. */
  const registerOn = (serverCtx) => {
    const rejected = (req, res) => {
      const connection = serverCtx.get('connection');
      const status = connection?.requestRejection?.(req);
      if (status === undefined) return false;
      res.statusCode = status;
      res.end();
      return true;
    };

    const subpathOf = (req, prefix) => new URL(String(req.url), 'http://localhost').pathname.slice(prefix.length);

    const readJsonBody = async (req, res) => {
      const essence = String(req.headers['content-type'] || '').split(';')[0].trim().toLowerCase();
      if (essence !== 'application/json') {
        sendJson(res, 415, { code: 1, msg: 'content-type must be application/json', data: null });
        return undefined;
      }
      let raw;
      try {
        raw = await readBoundedBody(req);
      } catch (err) {
        sendJson(res, 400, { code: 1, msg: `read body failed: ${err.message}`, data: null });
        return undefined;
      }
      if (raw === null) {
        sendJson(res, 413, { code: 1, msg: 'request body too large', data: null });
        return undefined;
      }
      try {
        return raw.length ? JSON.parse(raw) : {};
      } catch {
        sendJson(res, 400, { code: 1, msg: 'invalid JSON body', data: null });
        return undefined;
      }
    };

    const apiHandler = async (req, res) => {
      if (rejected(req, res)) return;
      if (req.method !== 'POST') { sendMethodNotAllowed(res, 'POST'); return; }
      const sub = subpathOf(req, GOSTC_API_PREFIX);
      if (!ALLOWED.has(sub)) {
        sendJson(res, 404, { code: 1, msg: `unknown gostc api: ${sub}`, data: null });
        return;
      }
      const payload = await readJsonBody(req, res);
      if (payload === undefined) return;

      let upstream;
      try {
        upstream = await fetch(`${backend}/api/${sub}`, {
          method: 'POST',
          headers: {
            'content-type': 'application/json',
            accept: 'application/json',
            ...(authHeader ? { authorization: authHeader } : {}),
          },
          body: JSON.stringify(payload),
        });
      } catch (err) {
        sendJson(res, 200, { code: 1, msg: `无法连接 gostc 后端（${backend}）：${err.message}`, data: null });
        return;
      }
      const text = await upstream.text();
      let envelope;
      try {
        envelope = JSON.parse(text);
      } catch {
        const hint = upstream.status === 401 || upstream.status === 403
          ? 'gostc 后端认证失败，请检查插件配置的用户名/密码'
          : `gostc 后端返回 ${upstream.status}`;
        envelope = { code: 1, msg: `${hint}：${text.slice(0, 200) || upstream.statusText}`, data: null };
      }
      sendJson(res, 200, envelope);
    };

    const processHandler = async (req, res) => {
      if (rejected(req, res)) return;
      const sub = subpathOf(req, GOSTC_PROCESS_PREFIX);
      if (!PROCESS_ALLOWED.has(sub)) {
        sendJson(res, 404, { code: 1, msg: `unknown gostc process op: ${sub}`, data: null });
        return;
      }
      const isReadGet = (sub === 'status' || sub === 'settings') && req.method === 'GET';
      if (req.method !== 'POST' && !isReadGet) {
        sendMethodNotAllowed(res, (sub === 'status' || sub === 'settings') ? 'GET, POST' : 'POST');
        return;
      }
      let payload = {};
      if (!isReadGet) {
        payload = await readJsonBody(req, res);
        if (payload === undefined) return;
      }
      if (sub === 'selftest') {
        try {
          const result = await manager.selftest();
          sendJson(res, 200, { code: 0, msg: 'ok', data: result });
        } catch (err) {
          sendJson(res, 200, { code: 1, msg: err.message || String(err), data: null });
        }
        return;
      }
      try {
        switch (sub) {
          case 'status': break;
          case 'settings': if (req.method === 'POST') await manager.updateSettings(payload); break;
          case 'start': await manager.start(); break;
          case 'stop': await manager.stop(); break;
          case 'restart': await manager.restart(); break;
          case 'download': await manager.ensureClient(!!payload.force); break;
          case 'service/install': await manager.serviceInstall(); break;
          case 'service/uninstall': await manager.serviceUninstall(); break;
          default: break;
        }
        sendJson(res, 200, { code: 0, msg: 'ok', data: manager.snapshot() });
      } catch (err) {
        sendJson(res, 200, { code: 1, msg: err.message || String(err), data: manager.snapshot() });
      }
    };

    return serverCtx.effect(() => {
      const disposers = [];
      const prefixes = serverCtx.webServer.prefixes;
      if (!prefixes || !prefixes.has(REGISTER_PATH)) {
        disposers.push(serverCtx.webServer.register({ kind: 'prefix', path: REGISTER_PATH, handler: apiHandler }));
      }
      if (!prefixes || !prefixes.has(REGISTER_PROCESS_PATH)) {
        disposers.push(serverCtx.webServer.register({ kind: 'prefix', path: REGISTER_PROCESS_PATH, handler: processHandler }));
      }
      return () => { for (const d of disposers) { try { d(); } catch {} } };
    }, 'gostc-routes');
  };

  // Activate immediately and bind the web server dynamically: a static inject
  // gate can leave the fiber pending if the service resolves in a later boot
  // phase. ctx.inject re-runs the callback whenever webServer (re)appears.
  if (ctx.get('webServer')) {
    diag('diag-3-bind.json', { stage: 'bind-direct', port: ctx.get('webServer').port ?? null });
    registerOn(ctx);
  } else {
    ctx.inject(['webServer'], (inner) => {
      diag('diag-3-bind.json', { stage: 'bind-inject', port: inner.webServer.port ?? null });
      registerOn(inner);
    });
  }
}