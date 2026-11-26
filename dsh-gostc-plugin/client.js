/**
 * Client half of @local/dsh-gostc.
 *
 * Registers a global management panel in the root-scoped `main` keyed slot and a
 * matching left-sidebar entry in `sidebar.panellist` (same id 'gostc'). The panel
 * manages gostc client connections, private tunnels, P2P links and runtime logs
 * through the Host proxy route (document-relative fetch 'gostc/api/...').
 *
 * Rules honored: React from the module table (no duplicate React); factories are
 * side-effect-free; resources live in ctx.effect; visible text goes through the
 * Client locale; controls are own markup under the `gostc-` prefix using only
 * --dsw-alias-* theme tokens; no iframe; no DOM outside the component.
 */
window.__ModuleLoader__.load({
  id: '@local/dsh-gostc',
  factory(require) {
    const React = require('react');
    const h = React.createElement;
    const { useState, useEffect, useCallback, useRef } = React;

    // ---- locale dictionaries (both baseline locales required) ----
    const MESSAGES = {
      zh: {
        panel: 'Gostc 穿透', title: 'Gostc 内网穿透', version: '版本',
        tabClient: '客户端', tabTunnel: '私有隧道', tabP2p: 'P2P 穿透', tabLogger: '运行日志',
        refresh: '刷新', add: '新增', edit: '编辑', delete: '删除', save: '保存',
        cancel: '取消', close: '关闭', retry: '重试', loading: '加载中…', empty: '暂无数据',
        running: '运行中', stopped: '已停止', autoStart: '自启动', server: '服务器',
        key: '密钥', name: '名称', address: '服务器地址', bind: '监听地址', port: '本地端口',
        tls: 'TLS', tlsOn: '启用 TLS', tlsOff: '禁用 TLS', localListen: '本地监听',
        confirmDeleteTitle: '确认删除', confirmDeleteMsg: '确定要删除该条目吗？此操作不可撤销。',
        createTitle: '新增', editTitle: '编辑', required: '请填写必填项',
        logTime: '时间', logType: '类型', logMessage: '内容', noLogs: '暂无日志',
        connectionFailed: '无法连接 gostc 后端', keyImmutable: '密钥创建后不可修改',
        on: '开', off: '关',
        tabProcess: '后端进程', procPhase: '状态', procMode: '模式', procVersion: '客户端版本', procTarget: '目标平台',
        phaseIdle: '空闲', phaseDownloading: '下载中', phaseStarting: '启动中', phaseRunning: '运行中',
        phaseStopped: '已停止', phaseError: '错误', phaseUnsupported: '不支持',
        modeManaged: '插件托管', modeExternal: '外部进程', modeService: '系统服务', modeNone: '无',
        start: '启动', stop: '停止', restart: '重启', downloadClient: '下载客户端', updateClient: '更新客户端',
        installService: '安装为服务', uninstallService: '卸载服务',
        platformUrlMissing: '未配置平台地址（platformUrl），无法自动下载客户端。请在插件配置中填写你的 gostc 平台地址。',
        unsupportedPlatform: '当前系统暂不支持自动安装客户端。',
        procLog: '进程日志', noProcLog: '暂无日志',
        procHint: '插件会自动下载并拉起 gostc 命令版客户端；若已有健康后端则直接复用。',
        busy: '处理中…',
      },
      en: {
        panel: 'Gostc', title: 'Gostc tunneling', version: 'Version',
        tabClient: 'Clients', tabTunnel: 'Tunnels', tabP2p: 'P2P', tabLogger: 'Logs',
        refresh: 'Refresh', add: 'Add', edit: 'Edit', delete: 'Delete', save: 'Save',
        cancel: 'Cancel', close: 'Close', retry: 'Retry', loading: 'Loading…', empty: 'No entries',
        running: 'Running', stopped: 'Stopped', autoStart: 'Auto start', server: 'Server',
        key: 'Key', name: 'Name', address: 'Server address', bind: 'Bind address', port: 'Local port',
        tls: 'TLS', tlsOn: 'Enable TLS', tlsOff: 'Disable TLS', localListen: 'Local listen',
        confirmDeleteTitle: 'Confirm delete', confirmDeleteMsg: 'Delete this entry? This cannot be undone.',
        createTitle: 'Add', editTitle: 'Edit', required: 'Please fill in the required fields',
        logTime: 'Time', logType: 'Type', logMessage: 'Message', noLogs: 'No logs',
        connectionFailed: 'Cannot reach the gostc backend', keyImmutable: 'The key cannot be changed after creation',
        on: 'On', off: 'Off',
        tabProcess: 'Backend', procPhase: 'Phase', procMode: 'Mode', procVersion: 'Client version', procTarget: 'Target',
        phaseIdle: 'Idle', phaseDownloading: 'Downloading', phaseStarting: 'Starting', phaseRunning: 'Running',
        phaseStopped: 'Stopped', phaseError: 'Error', phaseUnsupported: 'Unsupported',
        modeManaged: 'Plugin-managed', modeExternal: 'External process', modeService: 'Windows service', modeNone: 'None',
        start: 'Start', stop: 'Stop', restart: 'Restart', downloadClient: 'Download client', updateClient: 'Update client',
        installService: 'Install as service', uninstallService: 'Uninstall service',
        platformUrlMissing: 'No platform URL (platformUrl) configured, so the client cannot be downloaded automatically. Set your gostc platform URL in the plugin config.',
        unsupportedPlatform: 'This platform is not supported for automatic client installation yet.',
        procLog: 'Process log', noProcLog: 'No logs',
        procHint: 'The plugin downloads and launches the gostc command-line client automatically, reusing a healthy backend if one is already running.',
        busy: 'Working…',
      },
    };

    const KINDS = ['client', 'tunnel', 'p2p'];
    const TAB_LABEL = { client: 'tabClient', tunnel: 'tabTunnel', p2p: 'tabP2p', logger: 'tabLogger', process: 'tabProcess' };

    // ---- API helper: document-relative fetch through the Host proxy ----
    async function api(sub, body) {
      const res = await fetch('gostc/api/' + sub, {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify(body || {}),
      });
      let j;
      try {
        j = await res.json();
      } catch {
        throw new Error('后端响应解析失败 (' + res.status + ')');
      }
      if (j.code !== 0) throw new Error(j.msg || '请求失败');
      return j.data;
    }

    function defaultsFor(kind) {
      const base = { name: '', key: '', address: 'gostc.linux66.cn', tls: 1, autoStart: 0 };
      if (kind === 'client') return base;
      return Object.assign(base, { bind: '0.0.0.0', port: '8080' });
    }

    // ---- component-local stylesheet (theme tokens only) ----
    const CSS = `
.gostc-root{padding:var(--dsh-frame-top-clearance,0px) 24px 24px;color:var(--dsw-alias-label-primary);font-family:var(--dsw-font-family);height:100%;box-sizing:border-box;overflow:auto;}
.gostc-header{display:flex;align-items:baseline;gap:12px;margin-bottom:16px;}
.gostc-title{font-size:18px;font-weight:600;}
.gostc-version{color:var(--dsw-alias-label-tertiary);font-size:12px;}
.gostc-tabs{display:flex;gap:4px;border-bottom:var(--dsh-hairline-width,1px) solid var(--dsw-alias-border-l1);margin-bottom:16px;}
.gostc-tab{padding:8px 14px;cursor:pointer;border:none;background:transparent;color:var(--dsw-alias-label-secondary);font-size:14px;font-family:inherit;border-bottom:2px solid transparent;}
.gostc-tab[data-active="true"]{color:var(--dsw-alias-label-primary);border-bottom-color:var(--dsw-alias-state-business-primary);font-weight:600;}
.gostc-toolbar{display:flex;gap:8px;margin-bottom:16px;align-items:center;}
.gostc-spacer{flex:1;}
.gostc-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(280px,1fr));gap:12px;}
.gostc-card{background:var(--dsw-alias-card-fill);border:var(--dsh-hairline-width,1px) solid var(--dsw-alias-border-l1);border-radius:var(--dsw-radius-md);padding:14px 16px;display:flex;flex-direction:column;gap:8px;}
.gostc-card-head{display:flex;align-items:center;gap:10px;}
.gostc-card-name{font-weight:600;font-size:14px;flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;}
.gostc-row{display:flex;gap:6px;font-size:12px;color:var(--dsw-alias-label-secondary);}
.gostc-row-label{color:var(--dsw-alias-label-tertiary);min-width:60px;flex:none;}
.gostc-row-value{word-break:break-all;}
.gostc-card-actions{display:flex;gap:8px;margin-top:4px;align-items:center;}
.gostc-badge{font-size:11px;padding:1px 8px;border-radius:var(--dsw-radius-xs);background:var(--dsw-alias-interactive-bg-hover);color:var(--dsw-alias-label-secondary);}
.gostc-badge[data-on="true"]{color:var(--dsw-alias-state-success-primary);}
.gostc-btn{font-family:inherit;font-size:13px;padding:6px 14px;border-radius:var(--dsw-radius-sm);border:var(--dsh-hairline-width,1px) solid var(--dsw-alias-border-l2);background:var(--dsw-alias-button-elevated-fill);color:var(--dsw-alias-label-primary);cursor:pointer;}
.gostc-btn:hover{background:var(--dsw-alias-interactive-bg-hover);}
.gostc-btn:disabled{opacity:.5;cursor:not-allowed;}
.gostc-btn[data-variant="primary"]{background:var(--dsw-alias-button-info-fill);color:var(--dsw-alias-label-on-color);border-color:transparent;}
.gostc-btn[data-variant="primary"]:hover{background:var(--dsw-alias-button-info-hover);}
.gostc-btn[data-variant="danger"]{color:var(--dsw-alias-state-error-primary);}
.gostc-btn[data-variant="ghost"]{background:transparent;border-color:transparent;}
.gostc-btn:focus-visible{outline:var(--dsw-focus-ring-width,2px) solid var(--dsw-focus-ring-color,var(--dsw-alias-state-business-primary));outline-offset:1px;}
.gostc-switch{width:36px;height:20px;border-radius:10px;border:none;padding:2px;cursor:pointer;background:var(--dsw-alias-state-idle-primary);position:relative;transition:background var(--ds-transition-duration,.2s) var(--ds-ease-in-out,ease);flex:none;box-sizing:border-box;}
.gostc-switch[aria-checked="true"]{background:var(--dsw-alias-state-success-primary);}
.gostc-switch:disabled{opacity:.5;cursor:not-allowed;}
.gostc-switch-thumb{display:block;width:16px;height:16px;border-radius:50%;background:var(--dsw-alias-switch-thumb);transform:translateX(0);transition:transform var(--ds-transition-duration,.2s) var(--ds-ease-in-out,ease);}
.gostc-switch[aria-checked="true"] .gostc-switch-thumb{transform:translateX(16px);}
.gostc-mask{position:fixed;inset:0;background:var(--dsw-alias-bg-mask-1);display:flex;align-items:center;justify-content:center;z-index:1000;}
.gostc-modal{background:var(--dsw-alias-bg-layer-2);border:var(--dsh-hairline-width,1px) solid var(--dsw-alias-border-l2);border-radius:var(--dsw-radius-lg);width:min(440px,calc(100vw - 48px));max-height:calc(100vh - 96px);overflow:auto;padding:20px 22px;box-shadow:var(--dsw-shadow-lv2,0 8px 32px rgba(0,0,0,.2));}
.gostc-modal-title{font-size:16px;font-weight:600;margin-bottom:16px;}
.gostc-field{display:flex;flex-direction:column;gap:6px;margin-bottom:14px;}
.gostc-field-label{font-size:13px;color:var(--dsw-alias-label-secondary);}
.gostc-input,.gostc-select{font-family:inherit;font-size:13px;padding:7px 10px;border-radius:var(--dsw-radius-sm);border:var(--dsh-hairline-width,1px) solid var(--dsw-alias-border-l2);background:var(--dsw-specific-input-major);color:var(--dsw-alias-label-primary);width:100%;box-sizing:border-box;}
.gostc-input:focus,.gostc-select:focus{outline:var(--dsw-focus-ring-width,2px) solid var(--dsw-focus-ring-color,var(--dsw-alias-state-business-primary));outline-offset:0;}
.gostc-input:disabled{opacity:.6;}
.gostc-modal-footer{display:flex;justify-content:flex-end;gap:8px;margin-top:18px;}
.gostc-form-error{color:var(--dsw-alias-state-error-primary);font-size:12px;margin-bottom:8px;}
.gostc-banner{display:flex;align-items:center;gap:12px;padding:12px 16px;border-radius:var(--dsw-radius-md);background:var(--dsw-alias-interactive-bg-hover-danger);color:var(--dsw-alias-state-error-primary);font-size:13px;margin-bottom:16px;}
.gostc-empty{color:var(--dsw-alias-label-tertiary);font-size:13px;padding:40px 0;text-align:center;}
.gostc-muted{color:var(--dsw-alias-label-tertiary);font-size:13px;}
.gostc-logs{display:flex;flex-direction:column;gap:2px;font-family:var(--ds-font-family-code);font-size:12px;}
.gostc-log{display:flex;gap:10px;padding:4px 8px;border-radius:var(--dsw-radius-xs);}
.gostc-log:hover{background:var(--dsw-alias-interactive-bg-hover);}
.gostc-log-time{color:var(--dsw-alias-label-tertiary);flex:none;}
.gostc-log-type{flex:none;min-width:48px;color:var(--dsw-alias-label-secondary);}
.gostc-log-msg{word-break:break-all;color:var(--dsw-alias-label-primary);}
.gostc-proc-status{display:flex;flex-wrap:wrap;gap:10px 24px;background:var(--dsw-alias-card-fill);border:var(--dsh-hairline-width,1px) solid var(--dsw-alias-border-l1);border-radius:var(--dsw-radius-md);padding:14px 16px;margin-bottom:16px;}
.gostc-proc-item{display:flex;flex-direction:column;gap:4px;font-size:13px;}
.gostc-proc-item-label{color:var(--dsw-alias-label-tertiary);font-size:12px;}
.gostc-proc-item-value{color:var(--dsw-alias-label-primary);font-weight:600;}
.gostc-proc-item-value[data-tone="ok"]{color:var(--dsw-alias-state-success-primary);}
.gostc-proc-item-value[data-tone="err"]{color:var(--dsw-alias-state-error-primary);}
.gostc-proc-actions{display:flex;flex-wrap:wrap;gap:8px;margin-bottom:16px;}
.gostc-proc-log{background:var(--dsw-alias-bg-layer-2);border:var(--dsh-hairline-width,1px) solid var(--dsw-alias-border-l1);border-radius:var(--dsw-radius-md);padding:12px 14px;font-family:var(--ds-font-family-code);font-size:12px;color:var(--dsw-alias-label-secondary);white-space:pre-wrap;word-break:break-all;max-height:320px;overflow:auto;}
.gostc-proc-section-title{font-size:13px;font-weight:600;color:var(--dsw-alias-label-secondary);margin:8px 0;}
`;

    // ---- small controls ----
    function Button(props) {
      const { variant, children, ...rest } = props;
      return h('button', Object.assign({ type: 'button', className: 'gostc-btn', 'data-variant': variant || 'default' }, rest), children);
    }

    function Switch({ checked, onChange, disabled, label }) {
      return h('button', {
        type: 'button', className: 'gostc-switch', role: 'switch',
        'aria-checked': checked, 'aria-label': label, disabled,
        onClick: () => onChange && onChange(!checked),
      }, h('span', { className: 'gostc-switch-thumb' }));
    }

    function Modal({ title, onClose, children, footer }) {
      const ref = useRef(null);
      useEffect(() => {
        const prev = document.activeElement;
        const node = ref.current;
        if (node) node.focus();
        const onKey = (e) => { if (e.key === 'Escape') { e.stopPropagation(); onClose(); } };
        document.addEventListener('keydown', onKey, true);
        return () => {
          document.removeEventListener('keydown', onKey, true);
          if (prev && prev.focus) prev.focus();
        };
      }, [onClose]);
      return h('div', {
        className: 'gostc-mask',
        onMouseDown: (e) => { if (e.target === e.currentTarget) onClose(); },
      }, h('div', { className: 'gostc-modal', role: 'dialog', 'aria-modal': true, 'aria-label': title, tabIndex: -1, ref },
        h('div', { className: 'gostc-modal-title' }, title),
        children,
        footer ? h('div', { className: 'gostc-modal-footer' }, footer) : null));
    }

    function Field({ label, children }) {
      return h('label', { className: 'gostc-field' },
        h('span', { className: 'gostc-field-label' }, label), children);
    }

    // ---- entry form (create / edit) ----
    function EntryForm({ kind, mode, initial, t, onSubmit, onClose }) {
      const [values, setValues] = useState(() => Object.assign(defaultsFor(kind), initial || {}));
      const [error, setError] = useState(null);
      const [submitting, setSubmitting] = useState(false);
      const set = (k, v) => setValues((s) => Object.assign({}, s, { [k]: v }));
      const isTunnelLike = kind !== 'client';

      const submit = async () => {
        if (!values.name || !values.key || !values.address || (isTunnelLike && !values.port)) {
          setError(t('required'));
          return;
        }
        setSubmitting(true);
        setError(null);
        try {
          const payload = {
            name: values.name, key: values.key, address: values.address,
            tls: Number(values.tls), autoStart: Number(values.autoStart),
          };
          if (isTunnelLike) { payload.bind = values.bind || ''; payload.port = String(values.port); }
          await onSubmit(payload);
        } catch (err) {
          setError(err.message);
          setSubmitting(false);
        }
      };

      return h(Modal, {
        title: (mode === 'create' ? t('createTitle') : t('editTitle')) + ' · ' + t(TAB_LABEL[kind]),
        onClose,
        footer: [
          h(Button, { key: 'c', variant: 'ghost', onClick: onClose, disabled: submitting }, t('cancel')),
          h(Button, { key: 's', variant: 'primary', onClick: submit, disabled: submitting }, t('save')),
        ],
      },
        error ? h('div', { className: 'gostc-form-error' }, error) : null,
        h(Field, { label: t('name') }, h('input', { className: 'gostc-input', value: values.name, onChange: (e) => set('name', e.target.value) })),
        h(Field, { label: t('key') + (mode === 'edit' ? '（' + t('keyImmutable') + '）' : '') },
          h('input', { className: 'gostc-input', value: values.key, disabled: mode === 'edit', onChange: (e) => set('key', e.target.value) })),
        isTunnelLike ? h(Field, { label: t('bind') }, h('input', { className: 'gostc-input', value: values.bind, placeholder: '0.0.0.0', onChange: (e) => set('bind', e.target.value) })) : null,
        isTunnelLike ? h(Field, { label: t('port') }, h('input', { className: 'gostc-input', value: values.port, placeholder: '8080', onChange: (e) => set('port', e.target.value) })) : null,
        h(Field, { label: t('address') }, h('input', { className: 'gostc-input', value: values.address, onChange: (e) => set('address', e.target.value) })),
        h(Field, { label: t('tls') }, h('select', { className: 'gostc-select', value: values.tls, onChange: (e) => set('tls', Number(e.target.value)) },
          h('option', { value: 1 }, t('tlsOn')), h('option', { value: 2 }, t('tlsOff')))),
        h(Field, { label: t('autoStart') }, h(Switch, { checked: Number(values.autoStart) === 1, label: t('autoStart'), onChange: (v) => set('autoStart', v ? 1 : 0) })));
    }

    // ---- entry card ----
    function EntryCard({ kind, item, t, busy, onToggle, onEdit, onDelete }) {
      const running = item.status === 1;
      const isTunnelLike = kind !== 'client';
      const server = (item.tls === 1 ? 'wss://' : 'ws://') + item.address;
      return h('div', { className: 'gostc-card' },
        h('div', { className: 'gostc-card-head' },
          h('span', { className: 'gostc-card-name', title: item.name }, item.name || item.key),
          h(Switch, { checked: running, disabled: busy, label: running ? t('running') : t('stopped'), onChange: () => onToggle(item) })),
        h('div', { className: 'gostc-row' }, h('span', { className: 'gostc-row-label' }, t('server')), h('span', { className: 'gostc-row-value' }, server)),
        h('div', { className: 'gostc-row' }, h('span', { className: 'gostc-row-label' }, t('key')), h('span', { className: 'gostc-row-value' }, item.key)),
        isTunnelLike ? h('div', { className: 'gostc-row' }, h('span', { className: 'gostc-row-label' }, t('localListen')), h('span', { className: 'gostc-row-value' }, (item.bind || '0.0.0.0') + ':' + item.port)) : null,
        h('div', { className: 'gostc-card-actions' },
          h('span', { className: 'gostc-badge', 'data-on': item.autoStart === 1 }, t('autoStart') + ': ' + (item.autoStart === 1 ? t('on') : t('off'))),
          h('span', { className: 'gostc-spacer' }),
          h(Button, { variant: 'ghost', disabled: running || busy, onClick: () => onEdit(item) }, t('edit')),
          h(Button, { variant: 'danger', disabled: running || busy, onClick: () => onDelete(item) }, t('delete'))));
    }

    // ---- logger view ----
    function LoggerView({ t }) {
      const [logs, setLogs] = useState(null);
      const [error, setError] = useState(null);
      const load = useCallback(async () => {
        try { const data = await api('logger/list'); setLogs(data || []); setError(null); }
        catch (err) { setError(err.message); }
      }, []);
      useEffect(() => { load(); }, [load]);
      if (error) return h('div', { className: 'gostc-banner' }, h('span', { className: 'gostc-spacer' }, error), h(Button, { onClick: load }, t('retry')));
      if (!logs) return h('div', { className: 'gostc-empty' }, t('loading'));
      if (!logs.length) return h('div', { className: 'gostc-empty' }, t('noLogs'));
      return h('div', null,
        h('div', { className: 'gostc-toolbar' }, h('span', { className: 'gostc-spacer' }), h(Button, { onClick: load }, t('refresh'))),
        h('div', { className: 'gostc-logs' }, logs.map((log, i) => h('div', { className: 'gostc-log', key: i },
          h('span', { className: 'gostc-log-time' }, log.timestamp),
          h('span', { className: 'gostc-log-type' }, log.type),
          h('span', { className: 'gostc-log-msg' }, log.message)))));
    }

    // ---- backend / process control view ----
    const PHASE_LABEL = {
      idle: 'phaseIdle', downloading: 'phaseDownloading', starting: 'phaseStarting',
      running: 'phaseRunning', stopped: 'phaseStopped', error: 'phaseError', unsupported: 'phaseUnsupported',
    };
    const MODE_LABEL = { managed: 'modeManaged', external: 'modeExternal', service: 'modeService' };

    async function processCall(sub, body) {
      const res = await fetch('gostc/process/' + sub, {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify(body || {}),
      });
      return res.json();
    }

    function ProcessView({ t }) {
      const [snap, setSnap] = useState(null);
      const [error, setError] = useState(null);
      const [acting, setActing] = useState(false);

      const load = useCallback(async () => {
        try {
          const j = await processCall('status');
          if (j.data) setSnap(j.data);
          setError(j.code !== 0 ? j.msg : null);
        } catch (err) { setError(err.message); }
      }, []);

      const phase = snap && snap.phase;
      useEffect(() => {
        load();
        const id = setInterval(load, phase === 'running' ? 4000 : 8000);
        return () => clearInterval(id);
      }, [load, phase]);

      const act = async (sub, body) => {
        setActing(true);
        try {
          const j = await processCall(sub, body);
          if (j.data) setSnap(j.data);
          setError(j.code !== 0 ? j.msg : null);
        } catch (err) { setError(err.message); }
        finally { setActing(false); }
      };

      if (!snap) return h('div', { className: 'gostc-empty' }, t('loading'));

      const running = snap.phase === 'running';
      const tone = snap.phase === 'error' ? 'err' : (running ? 'ok' : '');
      const modeText = snap.mode ? t(MODE_LABEL[snap.mode]) : t('modeNone');
      const needPlatform = !snap.platformUrl && !snap.hasExe;

      return h('div', null,
        h('div', { className: 'gostc-muted', style: { marginBottom: '12px' } }, t('procHint')),
        !snap.supported ? h('div', { className: 'gostc-banner' }, h('span', { className: 'gostc-spacer' }, t('unsupportedPlatform'))) : null,
        needPlatform ? h('div', { className: 'gostc-banner' }, h('span', { className: 'gostc-spacer' }, t('platformUrlMissing'))) : null,
        error ? h('div', { className: 'gostc-banner' }, h('span', { className: 'gostc-spacer' }, error), h(Button, { onClick: load }, t('retry'))) : null,
        h('div', { className: 'gostc-proc-status' },
          h('div', { className: 'gostc-proc-item' }, h('span', { className: 'gostc-proc-item-label' }, t('procPhase')),
            h('span', { className: 'gostc-proc-item-value', 'data-tone': tone }, t(PHASE_LABEL[snap.phase] || 'phaseIdle'))),
          h('div', { className: 'gostc-proc-item' }, h('span', { className: 'gostc-proc-item-label' }, t('procMode')),
            h('span', { className: 'gostc-proc-item-value' }, modeText)),
          h('div', { className: 'gostc-proc-item' }, h('span', { className: 'gostc-proc-item-label' }, t('procVersion')),
            h('span', { className: 'gostc-proc-item-value' }, snap.version || '—')),
          h('div', { className: 'gostc-proc-item' }, h('span', { className: 'gostc-proc-item-label' }, t('procTarget')),
            h('span', { className: 'gostc-proc-item-value' }, snap.target ? (snap.target.os + '/' + snap.target.arch) : '—'))),
        h('div', { className: 'gostc-proc-actions' },
          h(Button, { variant: 'primary', disabled: acting || !snap.supported, onClick: () => act('start') }, acting ? t('busy') : t('start')),
          h(Button, { disabled: acting || snap.mode !== 'managed', onClick: () => act('stop') }, t('stop')),
          h(Button, { disabled: acting || !snap.supported, onClick: () => act('restart') }, t('restart')),
          h(Button, { disabled: acting || !snap.supported || !snap.platformUrl, onClick: () => act('download', { force: true }) }, snap.hasExe ? t('updateClient') : t('downloadClient')),
          h(Button, { disabled: acting || !snap.supported, onClick: () => act('service/install') }, t('installService')),
          h(Button, { variant: 'danger', disabled: acting || !snap.supported, onClick: () => act('service/uninstall') }, t('uninstallService'))),
        h('div', { className: 'gostc-proc-section-title' }, t('procLog')),
        snap.logTail ? h('div', { className: 'gostc-proc-log' }, snap.logTail) : h('div', { className: 'gostc-empty' }, t('noProcLog')));
    }

    // ---- main panel ----
    function GostcPage({ t }) {
      const [tab, setTab] = useState('client');
      const [data, setData] = useState({ client: null, tunnel: null, p2p: null });
      const [error, setError] = useState(null);
      const [loading, setLoading] = useState(false);
      const [version, setVersion] = useState('');
      const [modal, setModal] = useState(null);
      const [confirm, setConfirm] = useState(null);
      const [busy, setBusy] = useState({});

      const load = useCallback(async (kind) => {
        setLoading(true);
        try {
          const rows = await api(kind + '/list');
          setData((s) => Object.assign({}, s, { [kind]: rows || [] }));
          setError(null);
        } catch (err) {
          setError(err.message);
        } finally {
          setLoading(false);
        }
      }, []);

      useEffect(() => {
        api('config/query').then((d) => setVersion(d && d.version ? d.version : '')).catch(() => {});
      }, []);
      useEffect(() => { if (KINDS.includes(tab)) load(tab); }, [tab, load]);

      const setBusyKey = (key, on) => setBusy((s) => Object.assign({}, s, { [key]: on }));

      const toggle = async (kind, item) => {
        setBusyKey(item.key, true);
        try {
          await api(kind + '/status', { key: item.key, status: item.status === 1 ? 2 : 1 });
          await load(kind);
        } catch (err) { setError(err.message); }
        finally { setBusyKey(item.key, false); }
      };

      const submitForm = async (payload) => {
        const { kind, mode } = modal;
        await api(kind + '/' + (mode === 'create' ? 'create' : 'update'), payload);
        setModal(null);
        await load(kind);
      };

      const doDelete = async () => {
        const { kind, item } = confirm;
        setBusyKey(item.key, true);
        try {
          await api(kind + '/delete', { key: item.key });
          setConfirm(null);
          await load(kind);
        } catch (err) { setError(err.message); setConfirm(null); }
        finally { setBusyKey(item.key, false); }
      };

      const rows = data[tab];
      const isLogger = tab === 'logger';

      return h('div', { className: 'gostc-root' },
        h('style', null, CSS),
        h('div', { className: 'gostc-header' },
          h('span', { className: 'gostc-title' }, t('title')),
          version ? h('span', { className: 'gostc-version' }, t('version') + ' ' + version) : null),
        h('div', { className: 'gostc-tabs', role: 'tablist' },
          ['client', 'tunnel', 'p2p', 'logger', 'process'].map((k) => h('button', {
            key: k, type: 'button', className: 'gostc-tab', role: 'tab',
            'data-active': tab === k, 'aria-selected': tab === k, onClick: () => setTab(k),
          }, t(TAB_LABEL[k])))),
        tab === 'process' ? h(ProcessView, { t }) : isLogger ? h(LoggerView, { t }) : h(React.Fragment, null,
          error ? h('div', { className: 'gostc-banner' },
            h('span', { className: 'gostc-spacer' }, t('connectionFailed') + '：' + error),
            h(Button, { onClick: () => load(tab) }, t('retry'))) : null,
          h('div', { className: 'gostc-toolbar' },
            h('span', { className: 'gostc-muted' }, loading ? t('loading') : ''),
            h('span', { className: 'gostc-spacer' }),
            h(Button, { onClick: () => load(tab), disabled: loading }, t('refresh')),
            h(Button, { variant: 'primary', onClick: () => setModal({ kind: tab, mode: 'create', item: null }) }, t('add'))),
          !rows ? h('div', { className: 'gostc-empty' }, t('loading'))
            : rows.length === 0 ? h('div', { className: 'gostc-empty' }, t('empty'))
              : h('div', { className: 'gostc-grid' }, rows.map((item) => h(EntryCard, {
                key: item.key, kind: tab, item, t, busy: !!busy[item.key],
                onToggle: (it) => toggle(tab, it),
                onEdit: (it) => setModal({ kind: tab, mode: 'edit', item: it }),
                onDelete: (it) => setConfirm({ kind: tab, item: it }),
              })))),
        modal ? h(EntryForm, {
          kind: modal.kind, mode: modal.mode, initial: modal.item, t,
          onSubmit: submitForm, onClose: () => setModal(null),
        }) : null,
        confirm ? h(Modal, {
          title: t('confirmDeleteTitle'), onClose: () => setConfirm(null),
          footer: [
            h(Button, { key: 'c', variant: 'ghost', onClick: () => setConfirm(null) }, t('cancel')),
            h(Button, { key: 'd', variant: 'danger', onClick: doDelete }, t('delete')),
          ],
        }, h('div', { className: 'gostc-muted' }, t('confirmDeleteMsg'), ' ', h('strong', null, confirm.item.name || confirm.item.key))) : null);
    }

    // ---- sidebar icon (shell renders the row + selection automatically) ----
    function GostcIcon() {
      return h('svg', {
        viewBox: '0 0 24 24', width: 16, height: 16, fill: 'none',
        stroke: 'currentColor', strokeWidth: 1.7, strokeLinecap: 'round', strokeLinejoin: 'round',
        'aria-hidden': true, style: { display: 'block' },
      },
        h('path', { d: 'M4 16v-3a8 8 0 0 1 16 0v3' }),
        h('rect', { x: 2.5, y: 15.5, width: 4, height: 5, rx: 1.2 }),
        h('rect', { x: 17.5, y: 15.5, width: 4, height: 5, rx: 1.2 }),
        h('circle', { cx: 12, cy: 12.5, r: 1.4, fill: 'currentColor', stroke: 'none' }));
    }

    return {
      inject: ['slots', 'locale', 'layout'],
      apply(ctx) {
        ctx.effect(() => ctx.locale.register('gostc', MESSAGES), 'gostc: dictionary');
        const t = ctx.locale.bind('gostc');
        ctx.slots.inject('main', () => ctx.slots.register(
          { name: 'main', key: 'gostc', locale: 'gostc' },
          function GostcPanel(slotProps) { return h(GostcPage, Object.assign({}, slotProps, { t })); },
        ));
        ctx.slots.inject('sidebar.panellist', () => ctx.slots.register(
          { name: 'sidebar.panellist', id: 'gostc', order: 50, label: () => t('panel'), locale: 'gostc' },
          GostcIcon,
        ));
      },
    };
  },
});