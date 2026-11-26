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
        startDisabledRunning: '后端已在运行，无需再次启动。',
        stopDisabledHint: '停止仅对运行中的「插件托管 / 系统服务」可用；外部进程不由插件管理，未运行时也不可停止。',
        restartDisabledHint: '重启仅对运行中的「插件托管」进程可用。',
        installService: '安装为服务', uninstallService: '卸载服务',
        platformUrlMissing: '未配置平台地址（platformUrl），无法自动下载客户端。请在下方「设置」中填写你的 gostc 平台地址并保存。',
        unsupportedPlatform: '当前系统暂不支持自动安装客户端。',
        procLog: '进程日志', noProcLog: '暂无日志',
        procHint: '插件会自动下载并拉起 gostc 命令版客户端；若已有健康后端则直接复用。',
        settingsTitle: '设置', settingsHint: '在此填写你的 gostc 平台地址等信息，保存后立即生效并持久化（无需改配置文件）。',
        platformUrlLabel: '平台地址（platformUrl）', clientPackageLabel: '安装包文件名（留空自动按系统/架构挑选）',
        webAddrLabel: '客户端监听地址（webAddr）', webUserLabel: 'webui 用户名（可选）', webPwdLabel: 'webui 密码（可选）',
        autoStartLabel: '宿主启动时自动拉起客户端', saveSettings: '保存设置', settingsSaved: '设置已保存',
        tabAbout: '关于', aboutOfficial: '官网主页', aboutOfficialHint: '点击链接可直接跳转 gostc 官方平台（注册 / 登录 / 获取密钥 / 管理节点）。',
        aboutTagline: '在 DeepSeek Harness 里管理 gostc 内网穿透：自动下载并拉起客户端，管理客户端 / 隧道 / P2P 与日志。',
        aboutGuide: '快速上手', aboutLinks: '文档与源码', aboutContact: '联系与反馈', aboutEnv: '环境信息', aboutTips: '常见提示',
        guide1: '在「后端进程 → 设置」确认平台地址（默认 https://gostc.shop），按需改成你自己的平台并保存。',
        guide2: '点「下载客户端」按当前系统 / 架构自动安装命令版客户端，再点「启动」；若已有健康后端会自动复用，无需重复拉起。',
        guide3: '在「客户端」页新增连接：填名称、服务端地址与连接密钥（官网「客户端 → 连接密钥」获取），打开开关即建立穿透。',
        guide4: '需要把本地服务暴露到公网时，在「隧道」或「P2P」页新增条目并启动；公网地址见卡片上的「地址2」。',
        guide5: '运行日志在「日志」页查看；运行中的条目不可编辑 / 删除，请先停止再修改。',
        linkDocs: '使用文档（仓库 docs 目录）', linkPluginRepo: '本插件仓库（PonChunhui/dsh-gostc）',
        contactIssues: '提交问题 / 建议（本插件 Issues）',
        contactAdmin: '账号、节点、套餐等问题请联系你所用站点的管理员（官网登录后可见联系方式）；本插件仅负责本地客户端与穿透管理。',
        envPluginVer: '插件版本', envBackendVer: '后端版本', envTarget: '系统 / 架构', envDataDir: '数据目录',
        tip1: '连接失败：确认后端已启动、平台地址正确；面板会给出错误与重试按钮。',
        tip2: '运行中的客户端 / 隧道不可编辑或删除，先关闭开关停止后再修改。',
        tip3: '「安装为服务」会触发 Windows UAC 提权；取消授权会报错，属正常现象。',
        tip4: '客户端连接密钥等同凭证，请勿外传或截图公开；泄露可在官网重置。',
        deepRegister: '注册 / 登录', deepKey: '获取连接密钥', deepDownload: '下载客户端', deepTunnel: '管理隧道',
        aboutSelftest: '状态自检', selftestRun: '一键自检', selftestBackend: '本地后端', selftestPlatform: '平台可达',
        selftestOk: '正常', selftestFail: '异常',
        selftestHintOk: '后端与平台均正常，可直接在「客户端」页添加连接并开启隧道。',
        selftestHintBad: '后端未就绪：到「后端进程」页启动或下载客户端；平台不可达：检查网络 / 平台地址。',
        selftestIdle: '点击「立即自检」探测本地后端与平台连通性，并给出下一步建议。',
        aboutFaq: '常见问题',
        faq1q: '提示「无法连接 gostc 后端」怎么办？', faq1a: '到「后端进程」页点「启动」或「下载客户端」；若已有 gostc 客户端在运行，确认其 web 监听地址与设置中的 webAddr 一致（默认 127.0.0.1:18080）。',
        faq2q: '隧道 / 客户端开关打开后不生效？', faq2a: '确认连接密钥与服务器地址正确、服务端节点在线；查看「日志」页的 FRP 报错；运行中的条目需先停止才能修改配置。',
        faq3q: '连接密钥在哪里获取？', faq3a: '点上方「获取连接密钥」直达官网客户端页，创建或查看客户端即可看到密钥；密钥等同凭证，请勿外传。',
        faq4q: '如何让 gostc 开机自启、不依赖本插件？', faq4a: '在「后端进程」页点「安装为服务」（Windows UAC 提权），之后 gostc 作为系统服务独立运行；卸载服务可恢复由插件托管。',
        linkChangelog: '更新日志', linkRoadmap: '路线图',
        contactEmail: '邮箱', contactQQ: 'QQ 群', contactWeChat: '微信群',
        contactEmailLabel: '联系邮箱（可选）', contactQQLabel: 'QQ 群号（可选）', contactWeChatLabel: '微信群二维码图片 URL（可选）',
        diagCopy: '复制诊断信息', diagCopied: '已复制，可粘贴到 Issues',
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
        startDisabledRunning: 'Backend is already running; no need to start again.',
        stopDisabledHint: 'Stop is only available for a running plugin-managed child or Windows service; external processes are not managed here, and nothing can be stopped while idle.',
        restartDisabledHint: 'Restart is only available for a running plugin-managed child.',
        installService: 'Install as service', uninstallService: 'Uninstall service',
        platformUrlMissing: 'No platform URL (platformUrl) configured, so the client cannot be downloaded automatically. Set your gostc platform URL in the plugin config.',
        unsupportedPlatform: 'This platform is not supported for automatic client installation yet.',
        procLog: 'Process log', noProcLog: 'No logs',
        procHint: 'The plugin downloads and launches the gostc command-line client automatically, reusing a healthy backend if one is already running.',
        settingsTitle: 'Settings', settingsHint: 'Enter your gostc platform URL and related options here; saving takes effect immediately and persists (no config-file editing needed).',
        platformUrlLabel: 'Platform URL (platformUrl)', clientPackageLabel: 'Package filename (empty = auto-pick by OS/arch)',
        webAddrLabel: 'Client listen address (webAddr)', webUserLabel: 'webui username (optional)', webPwdLabel: 'webui password (optional)',
        autoStartLabel: 'Auto-launch client on host start', saveSettings: 'Save settings', settingsSaved: 'Settings saved',
        tabAbout: 'About', aboutOfficial: 'Official website', aboutOfficialHint: 'Click to open the gostc platform (sign up / sign in / get keys / manage nodes).',
        aboutTagline: 'Manage gostc intranet penetration inside DeepSeek Harness: auto-download and launch the client, manage clients / tunnels / P2P and logs.',
        aboutGuide: 'Quick start', aboutLinks: 'Docs & source', aboutContact: 'Contact & feedback', aboutEnv: 'Environment', aboutTips: 'Tips',
        guide1: 'In Backend → Settings confirm the platform URL (default https://gostc.shop); change it to your own platform if needed and save.',
        guide2: 'Click Download client to auto-install the CLI for your OS/arch, then Start. A healthy existing backend is reused automatically.',
        guide3: 'On the Clients tab add a connection: name, server address and key (from the website "Clients → key"); flip the switch to tunnel.',
        guide4: 'To expose a local service publicly, add an entry on the Tunnel or P2P tab and start it; the public address is shown as "address 2".',
        guide5: 'View runtime logs on the Logs tab; running entries cannot be edited/deleted — stop them first.',
        linkDocs: 'User docs (repo docs folder)', linkPluginRepo: 'This plugin repo (PonChunhui/dsh-gostc)',
        contactIssues: 'Report issues / suggestions (plugin Issues)',
        contactAdmin: 'For account / node / plan issues contact your site admin (contact info on the website after sign-in); this plugin only manages the local client and tunnels.',
        envPluginVer: 'Plugin version', envBackendVer: 'Backend version', envTarget: 'OS / arch', envDataDir: 'Data dir',
        tip1: 'Connection failed: make sure the backend is running and the platform URL is correct; the panel shows the error and a retry button.',
        tip2: 'Running clients / tunnels cannot be edited or deleted; switch them off first.',
        tip3: 'Install as service triggers Windows UAC elevation; cancelling it reports an error, which is expected.',
        tip4: 'The client key is a credential — do not share or screenshot it publicly; reset it on the website if leaked.',
        deepRegister: 'Sign up / sign in', deepKey: 'Get connection key', deepDownload: 'Download client', deepTunnel: 'Manage tunnels',
        aboutSelftest: 'Self test', selftestRun: 'Run self test', selftestBackend: 'Local backend', selftestPlatform: 'Platform reachable',
        selftestOk: 'OK', selftestFail: 'Failed',
        selftestHintOk: 'Backend and platform are healthy — add a connection on the Clients tab and enable tunnels.',
        selftestHintBad: 'Backend not ready: start or download the client on the Backend tab. Platform unreachable: check network / platform URL.',
        selftestIdle: 'Run the self test to probe local backend and platform connectivity, with next-step guidance.',
        aboutFaq: 'FAQ',
        faq1q: 'What if it says "cannot connect to gostc backend"?', faq1a: 'On the Backend tab click Start or Download client. If a gostc client is already running, make sure its web listen address matches webAddr in Settings (default 127.0.0.1:18080).',
        faq2q: 'A tunnel / client switch is on but nothing happens?', faq2a: 'Verify the connection key and server address, and that the server node is online; check FRP errors on the Logs tab. Running entries must be stopped before editing.',
        faq3q: 'Where do I get the connection key?', faq3a: 'Click "Get connection key" above to open the official client page; create or view a client to see its key. Treat the key as a credential.',
        faq4q: 'How to auto-start gostc at boot, independent of this plugin?', faq4a: 'On the Backend tab click "Install as service" (Windows UAC). gostc then runs as a system service; uninstall the service to return to plugin-managed mode.',
        linkChangelog: 'Changelog', linkRoadmap: 'Roadmap',
        contactEmail: 'Email', contactQQ: 'QQ group', contactWeChat: 'WeChat group',
        contactEmailLabel: 'Contact email (optional)', contactQQLabel: 'QQ group number (optional)', contactWeChatLabel: 'WeChat QR image URL (optional)',
        diagCopy: 'Copy diagnostics', diagCopied: 'Copied — paste into an Issue',
        busy: 'Working…',
      },
    };

    const KINDS = ['client', 'tunnel', 'p2p'];
    const TAB_LABEL = { client: 'tabClient', tunnel: 'tabTunnel', p2p: 'tabP2p', logger: 'tabLogger', process: 'tabProcess', about: 'tabAbout' };
    const OFFICIAL_URL = 'https://gostc.shop';
    const UPSTREAM_REPO = 'https://github.com/SianHH/gostc-open';
    const PLUGIN_REPO = 'https://github.com/PonChunhui/dsh-gostc';
    const PLUGIN_VERSION = '1.4.5';

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
      const base = { name: '', key: '', address: 'gostc.shop', tls: 1, autoStart: 0 };
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
.gostc-proc-settings{background:var(--dsw-alias-card-fill);border:var(--dsh-hairline-width,1px) solid var(--dsw-alias-border-l1);border-radius:var(--dsw-radius-md);padding:14px 16px;margin-bottom:16px;}
.gostc-proc-settings-head{display:flex;align-items:center;gap:8px;cursor:pointer;user-select:none;}
.gostc-proc-settings-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(240px,1fr));gap:0 16px;margin-top:12px;}
.gostc-proc-settings-foot{display:flex;align-items:center;gap:12px;margin-top:4px;}
.gostc-set-ok{font-size:12px;color:var(--dsw-alias-state-success-primary);}
.gostc-about-official{display:flex;flex-direction:column;gap:6px;margin-top:8px;padding:12px;border:var(--dsh-hairline-width) solid var(--dsw-alias-border-l1);border-radius:var(--dsw-radius-md);background:var(--dsw-alias-card-fill);}
.gostc-about-link{display:inline-flex;align-items:center;gap:6px;color:var(--dsw-alias-state-business-primary);text-decoration:none;font-weight:600;font-size:13px;}
.gostc-about-link:hover{text-decoration:underline;}
.gostc-about-links{display:flex;flex-direction:column;gap:8px;margin-top:8px;}
.gostc-faq-item{border:1px solid var(--dsw-alias-border-l1);border-radius:var(--dsw-radius-sm);margin-top:8px;overflow:hidden;}
.gostc-faq-q{display:flex;justify-content:space-between;align-items:center;gap:8px;padding:10px 12px;cursor:pointer;font-weight:600;font-size:13px;}
.gostc-faq-q:hover{background:var(--dsw-alias-interactive-bg-hover);}
.gostc-faq-a{padding:0 12px 10px;font-size:13px;color:var(--dsw-alias-label-secondary);line-height:1.6;}
.gostc-dot{display:inline-block;width:8px;height:8px;border-radius:50%;flex:none;}
.gostc-qr{width:120px;height:120px;object-fit:contain;border:1px solid var(--dsw-alias-border-l1);border-radius:var(--dsw-radius-sm);background:#fff;}
.gostc-about-list{margin:8px 0 0;padding-left:20px;display:flex;flex-direction:column;gap:6px;color:var(--dsw-alias-label-secondary);font-size:13px;line-height:1.6;}
.gostc-about{display:flex;flex-direction:column;gap:16px;}
.gostc-about-hero{display:flex;flex-wrap:wrap;align-items:center;gap:16px;padding:16px 18px;border:var(--dsh-hairline-width,1px) solid var(--dsw-alias-border-l1);border-radius:var(--dsw-radius-lg);background:var(--dsw-alias-card-fill);}
.gostc-about-hero-main{display:flex;flex-direction:column;gap:6px;flex:1;min-width:220px;}
.gostc-about-title{display:flex;align-items:center;gap:8px;font-size:18px;font-weight:700;color:var(--dsw-alias-label-primary);}
.gostc-about-ver{font-size:11px;font-weight:600;padding:2px 8px;border-radius:var(--dsw-radius-xs);background:var(--dsw-alias-interactive-bg-hover);color:var(--dsw-alias-label-secondary);}
.gostc-about-tag{font-size:13px;color:var(--dsw-alias-label-secondary);line-height:1.5;}
.gostc-about-quick{display:flex;flex-wrap:wrap;gap:8px;}
.gostc-about-chip{display:inline-flex;align-items:center;gap:6px;padding:6px 12px;border:var(--dsh-hairline-width,1px) solid var(--dsw-alias-border-l2);border-radius:var(--dsw-radius-sm);background:var(--dsw-alias-button-elevated-fill);color:var(--dsw-alias-label-primary);font-size:12px;font-weight:600;text-decoration:none;}
.gostc-about-chip:hover{background:var(--dsw-alias-interactive-bg-hover);border-color:var(--dsw-alias-state-business-primary);color:var(--dsw-alias-state-business-primary);}
.gostc-about-cols{display:grid;grid-template-columns:minmax(0,1.15fr) minmax(0,.85fr);gap:16px;align-items:start;}
@media (max-width:900px){.gostc-about-cols{grid-template-columns:1fr;}}
.gostc-about-col{display:flex;flex-direction:column;gap:16px;min-width:0;}
.gostc-about-card{display:flex;flex-direction:column;gap:10px;padding:14px 16px;border:var(--dsh-hairline-width,1px) solid var(--dsw-alias-border-l1);border-radius:var(--dsw-radius-md);background:var(--dsw-alias-card-fill);}
.gostc-about-card-title{display:flex;align-items:center;justify-content:space-between;gap:8px;margin:0;font-size:13px;font-weight:600;color:var(--dsw-alias-label-secondary);}
.gostc-about-card .gostc-about-links{margin-top:0;}
.gostc-about-card .gostc-about-list{margin-top:0;}
.gostc-about-card .gostc-faq-item{margin-top:0;}
.gostc-about-card .gostc-faq-item + .gostc-faq-item{margin-top:8px;}
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

    function SettingsForm({ t, snap, onSaved }) {
      const [open, setOpen] = useState(!snap.platformUrl && !snap.hasExe);
      const [form, setForm] = useState(() => ({
        platformUrl: snap.platformUrl || '',
        clientPackage: snap.clientPackage || '',
        webAddr: snap.webAddr || '127.0.0.1:18080',
        webUser: snap.webUser || '',
        webPwd: snap.webPwd || '',
        autoStart: snap.autoStart !== false,
        contactEmail: snap.contactEmail || '',
        contactQQ: snap.contactQQ || '',
        contactWeChat: snap.contactWeChat || '',
      }));
      const [saving, setSaving] = useState(false);
      const [saved, setSaved] = useState(false);
      const [err, setErr] = useState(null);
      const set = (k, v) => { setForm((f) => Object.assign({}, f, { [k]: v })); setSaved(false); };
      const save = async () => {
        setSaving(true); setErr(null);
        try {
          const j = await processCall('settings', form);
          if (j.code !== 0) { setErr(j.msg); }
          else { setSaved(true); if (j.data) onSaved(j.data); }
        } catch (e) { setErr(e.message); }
        finally { setSaving(false); }
      };
      return h('div', { className: 'gostc-proc-settings' },
        h('div', { className: 'gostc-proc-settings-head', onClick: () => setOpen(!open), role: 'button', tabIndex: 0,
          onKeyDown: (e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); setOpen(!open); } } },
          h('span', { className: 'gostc-proc-section-title', style: { margin: 0 } }, t('settingsTitle')),
          h('span', { className: 'gostc-muted' }, open ? '▾' : '▸')),
        open ? h(React.Fragment, null,
          h('div', { className: 'gostc-muted', style: { marginTop: '8px' } }, t('settingsHint')),
          h('div', { className: 'gostc-proc-settings-grid' },
            h(Field, { label: t('platformUrlLabel') }, h('input', { className: 'gostc-input', value: form.platformUrl, placeholder: 'https://your-gostc-server', onChange: (e) => set('platformUrl', e.target.value) })),
            h(Field, { label: t('clientPackageLabel') }, h('input', { className: 'gostc-input', value: form.clientPackage, placeholder: 'gostc-cli.exe', onChange: (e) => set('clientPackage', e.target.value) })),
            h(Field, { label: t('webAddrLabel') }, h('input', { className: 'gostc-input', value: form.webAddr, onChange: (e) => set('webAddr', e.target.value) })),
            h(Field, { label: t('webUserLabel') }, h('input', { className: 'gostc-input', value: form.webUser, onChange: (e) => set('webUser', e.target.value) })),
            h(Field, { label: t('webPwdLabel') }, h('input', { className: 'gostc-input', type: 'password', value: form.webPwd, onChange: (e) => set('webPwd', e.target.value) })),
            h(Field, { label: t('autoStartLabel') }, h(Switch, { checked: form.autoStart, onChange: (v) => set('autoStart', v) })),
            h(Field, { label: t('contactEmailLabel') }, h('input', { className: 'gostc-input', value: form.contactEmail, placeholder: 'support@example.com', onChange: (e) => set('contactEmail', e.target.value) })),
            h(Field, { label: t('contactQQLabel') }, h('input', { className: 'gostc-input', value: form.contactQQ, placeholder: '123456789', onChange: (e) => set('contactQQ', e.target.value) })),
            h(Field, { label: t('contactWeChatLabel') }, h('input', { className: 'gostc-input', value: form.contactWeChat, placeholder: 'https://…/qrcode.png', onChange: (e) => set('contactWeChat', e.target.value) }))),
          h('div', { className: 'gostc-proc-settings-foot' },
            h(Button, { variant: 'primary', disabled: saving, onClick: save }, saving ? t('busy') : t('saveSettings')),
            saved ? h('span', { className: 'gostc-set-ok' }, t('settingsSaved')) : null,
            err ? h('span', { className: 'gostc-muted', style: { color: 'var(--dsw-alias-state-error-primary)' } }, err) : null)) : null);
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
      const canStop = running && (snap.mode === 'managed' || snap.mode === 'service');
      const canRestart = running && snap.mode === 'managed';
      const tone = snap.phase === 'error' ? 'err' : (running ? 'ok' : '');
      const modeText = snap.mode ? t(MODE_LABEL[snap.mode]) : t('modeNone');
      const needPlatform = !snap.platformUrl && !snap.hasExe;

      return h('div', null,
        h('div', { className: 'gostc-muted', style: { marginBottom: '12px' } }, t('procHint')),
        !snap.supported ? h('div', { className: 'gostc-banner' }, h('span', { className: 'gostc-spacer' }, t('unsupportedPlatform'))) : null,
        needPlatform ? h('div', { className: 'gostc-banner' }, h('span', { className: 'gostc-spacer' }, t('platformUrlMissing'))) : null,
        error ? h('div', { className: 'gostc-banner' }, h('span', { className: 'gostc-spacer' }, error), h(Button, { onClick: load }, t('retry'))) : null,
        h(SettingsForm, { t, snap, onSaved: setSnap }),
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
          h(Button, { variant: 'primary', disabled: acting || !snap.supported || running, title: running ? t('startDisabledRunning') : undefined, onClick: () => act('start') }, acting ? t('busy') : t('start')),
          h(Button, { disabled: acting || !canStop, title: canStop ? undefined : t('stopDisabledHint'), onClick: () => act('stop') }, t('stop')),
          h(Button, { disabled: acting || !canRestart, title: canRestart ? undefined : t('restartDisabledHint'), onClick: () => act('restart') }, t('restart')),
          h(Button, { disabled: acting || !snap.supported || !snap.platformUrl, onClick: () => act('download', { force: true }) }, snap.hasExe ? t('updateClient') : t('downloadClient')),
          h(Button, { disabled: acting || !snap.supported, onClick: () => act('service/install') }, t('installService')),
          h(Button, { variant: 'danger', disabled: acting || !snap.supported, onClick: () => act('service/uninstall') }, t('uninstallService'))),
        h('div', { className: 'gostc-proc-section-title' }, t('procLog')),
        snap.logTail ? h('div', { className: 'gostc-proc-log' }, snap.logTail) : h('div', { className: 'gostc-empty' }, t('noProcLog')));
    }

    // ---- main panel ----
    function AboutLink({ href, children }) {
      return h('a', { className: 'gostc-about-link', href, target: '_blank', rel: 'noopener noreferrer' }, children, h('span', { 'aria-hidden': true }, '↗'));
    }

    function FaqItem({ q, a }) {
      const [open, setOpen] = useState(false);
      return h('div', { className: 'gostc-faq-item' },
        h('div', { className: 'gostc-faq-q', role: 'button', tabIndex: 0, onClick: () => setOpen(!open),
          onKeyDown: (e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); setOpen(!open); } } },
          h('span', null, q), h('span', { className: 'gostc-muted' }, open ? '▾' : '▸')),
        open ? h('div', { className: 'gostc-faq-a' }, a) : null);
    }

    function copyText(text) {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        return navigator.clipboard.writeText(text).then(() => true).catch(() => legacyCopy(text));
      }
      return Promise.resolve(legacyCopy(text));
    }
    function legacyCopy(text) {
      try {
        const ta = document.createElement('textarea');
        ta.value = text; ta.style.position = 'fixed'; ta.style.opacity = '0';
        document.body.appendChild(ta); ta.select();
        const ok = document.execCommand('copy');
        document.body.removeChild(ta);
        return ok;
      } catch { return false; }
    }

    function AboutView({ t, version }) {
      const [proc, setProc] = useState(null);
      const [test, setTest] = useState(null);
      const [testing, setTesting] = useState(false);
      const [copied, setCopied] = useState(false);
      useEffect(() => {
        let alive = true;
        processCall('status', {}).then((j) => { if (alive && j && j.data) setProc(j.data); }).catch(() => {});
        return () => { alive = false; };
      }, []);
      const card = (title, action, body) => h('div', { className: 'gostc-about-card' },
        h('div', { className: 'gostc-about-card-title' }, h('span', null, title), action || null), body);
      const chip = (href, label) => h('a', { className: 'gostc-about-chip', href, target: '_blank', rel: 'noopener noreferrer' },
        label, h('span', { 'aria-hidden': true }, '↗'));
      const dot = (ok) => h('span', { className: 'gostc-dot', style: { background: ok ? 'var(--dsw-alias-state-success-primary)' : 'var(--dsw-alias-state-error-primary)' } });
      const runTest = async () => {
        setTesting(true);
        try { const j = await processCall('selftest', {}); setTest(j && j.code === 0 ? j.data : { error: j && j.msg }); }
        catch (e) { setTest({ error: e.message }); }
        finally { setTesting(false); }
      };
      const diagText = () => [
        'gostc plugin diagnostics',
        'plugin: dsh-gostc ' + PLUGIN_VERSION,
        'backend version: ' + (version || '-'),
        'backend url: ' + (proc ? proc.backend : '-'),
        'phase: ' + (proc ? proc.phase : '-') + ' / mode: ' + (proc ? (proc.mode || '-') : '-'),
        'target: ' + (proc && proc.target ? proc.target.os + '/' + proc.target.arch : '-'),
        'dataDir: ' + (proc ? proc.dataDir : '-'),
        'platformUrl: ' + (proc ? proc.platformUrl : '-'),
        'time: ' + new Date().toISOString(),
      ].join('\n');
      const doCopy = async () => { const ok = await copyText(diagText()); setCopied(ok); setTimeout(() => setCopied(false), 2000); };
      return h('div', { className: 'gostc-about' },
        h('div', { className: 'gostc-about-hero' },
          h('div', { className: 'gostc-about-hero-main' },
            h('div', { className: 'gostc-about-title' }, 'Gostc',
              h('span', { className: 'gostc-about-ver' }, 'v' + PLUGIN_VERSION)),
            h('div', { className: 'gostc-about-tag' }, t('aboutTagline')),
            h(AboutLink, { href: OFFICIAL_URL }, OFFICIAL_URL)),
          h('div', { className: 'gostc-about-quick' },
            chip(OFFICIAL_URL + '/login', t('deepRegister')),
            chip(OFFICIAL_URL + '/normal/gost/client', t('deepKey')),
            chip(OFFICIAL_URL + '/download', t('deepDownload')),
            chip(OFFICIAL_URL + '/normal/gost/client/tunnel', t('deepTunnel')))),
        h('div', { className: 'gostc-about-cols' },
          h('div', { className: 'gostc-about-col' },
            card(t('aboutGuide'), null,
              h('ol', { className: 'gostc-about-list' },
                ['guide1', 'guide2', 'guide3', 'guide4', 'guide5'].map((k) => h('li', { key: k }, t(k))))),
            card(t('aboutFaq'), null,
              h('div', null, [1, 2, 3, 4].map((n) => h(FaqItem, { key: n, q: t('faq' + n + 'q'), a: t('faq' + n + 'a') })))),
            card(t('aboutTips'), null,
              h('ul', { className: 'gostc-about-list' },
                ['tip1', 'tip2', 'tip3', 'tip4'].map((k) => h('li', { key: k }, t(k)))))),
          h('div', { className: 'gostc-about-col' },
            card(t('aboutSelftest'),
              h(Button, { variant: 'primary', disabled: testing, onClick: runTest }, testing ? t('busy') : t('selftestRun')),
              test ? h('div', null,
                h('div', { className: 'gostc-proc-grid' },
                  h('div', { className: 'gostc-proc-item' }, dot(test.backend && test.backend.healthy), h('span', { className: 'gostc-proc-k' }, t('selftestBackend')), h('span', { className: 'gostc-proc-v' }, test.backend && test.backend.healthy ? (t('selftestOk') + (test.backend.version ? ' ' + test.backend.version : '')) : t('selftestFail'))),
                  h('div', { className: 'gostc-proc-item' }, dot(test.platform && test.platform.reachable), h('span', { className: 'gostc-proc-k' }, t('selftestPlatform')), h('span', { className: 'gostc-proc-v' }, test.platform && test.platform.reachable ? t('selftestOk') : (test.platform && test.platform.error) || t('selftestFail'))),
                  test.error ? h('div', { className: 'gostc-proc-item' }, h('span', { className: 'gostc-proc-v', style: { color: 'var(--dsw-alias-state-error-primary)' } }, test.error)) : null),
                h('div', { className: 'gostc-muted' }, test.backend && test.backend.healthy ? t('selftestHintOk') : t('selftestHintBad')))
                : h('div', { className: 'gostc-muted' }, t('selftestIdle'))),
            card(t('aboutEnv'),
              h(Button, { onClick: doCopy }, copied ? t('diagCopied') : t('diagCopy')),
              h('div', { className: 'gostc-proc-grid' },
                h('div', { className: 'gostc-proc-item' }, h('span', { className: 'gostc-proc-k' }, t('envPluginVer')), h('span', { className: 'gostc-proc-v' }, PLUGIN_VERSION)),
                h('div', { className: 'gostc-proc-item' }, h('span', { className: 'gostc-proc-k' }, t('envBackendVer')), h('span', { className: 'gostc-proc-v' }, version || '—')),
                h('div', { className: 'gostc-proc-item' }, h('span', { className: 'gostc-proc-k' }, t('envTarget')), h('span', { className: 'gostc-proc-v' }, proc && proc.target ? proc.target.os + '/' + proc.target.arch : '—')),
                h('div', { className: 'gostc-proc-item' }, h('span', { className: 'gostc-proc-k' }, t('envDataDir')), h('span', { className: 'gostc-proc-v', style: { wordBreak: 'break-all' } }, proc ? proc.dataDir : '—')))),
            card(t('aboutContact'), null,
              h('div', { className: 'gostc-about-links' },
                h(AboutLink, { href: PLUGIN_REPO + '/issues' }, t('contactIssues')),
                proc && proc.contactEmail ? h(AboutLink, { href: 'mailto:' + proc.contactEmail }, t('contactEmail') + ': ' + proc.contactEmail) : null,
                proc && proc.contactQQ ? h('span', { className: 'gostc-about-link' }, t('contactQQ') + ': ' + proc.contactQQ) : null,
                proc && proc.contactWeChat ? h('img', { className: 'gostc-qr', src: proc.contactWeChat, alt: t('contactWeChat') }) : null,
                h('div', { className: 'gostc-muted' }, t('contactAdmin')))),
            card(t('aboutLinks'), null,
              h('div', { className: 'gostc-about-links' },
                h(AboutLink, { href: UPSTREAM_REPO + '/tree/main/docs' }, t('linkDocs')),
                h(AboutLink, { href: PLUGIN_REPO }, t('linkPluginRepo')),
                h(AboutLink, { href: PLUGIN_REPO + '/blob/main/CHANGELOG.md' }, t('linkChangelog')),
                h(AboutLink, { href: PLUGIN_REPO + '/blob/main/ROADMAP.md' }, t('linkRoadmap')))))));
    }

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
          ['client', 'tunnel', 'p2p', 'logger', 'process', 'about'].map((k) => h('button', {
            key: k, type: 'button', className: 'gostc-tab', role: 'tab',
            'data-active': tab === k, 'aria-selected': tab === k, onClick: () => setTab(k),
          }, t(TAB_LABEL[k])))),
        tab === 'about' ? h(AboutView, { t, version }) : tab === 'process' ? h(ProcessView, { t }) : isLogger ? h(LoggerView, { t }) : h(React.Fragment, null,
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