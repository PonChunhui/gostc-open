/**
 * Browser-safe route constants shared by the Host proxy and the Client panel.
 * No Node imports: this module is loaded in the browser module table.
 *
 * The Host registers a prefix route at GOSTC_API_PREFIX. The browser fetches the
 * DOCUMENT-RELATIVE form (no leading slash) so the request resolves against the
 * document base in Web and is forwarded over the Desktop file:// IPC fetch bridge.
 */
export const GOSTC_API_PREFIX = '/gostc/api/';

/** Document-relative route for a gostc API subpath, e.g. gostcRoute('tunnel/list'). */
export const gostcRoute = (sub) => GOSTC_API_PREFIX.slice(1) + sub;

/** Allowed upstream subpaths (also enforced Host-side as an anti-SSRF allowlist). */
export const GOSTC_SUBPATHS = [
  'client/create', 'client/update', 'client/delete', 'client/list', 'client/status',
  'tunnel/create', 'tunnel/update', 'tunnel/delete', 'tunnel/list', 'tunnel/status',
  'p2p/create', 'p2p/update', 'p2p/delete', 'p2p/list', 'p2p/status',
  'logger/list', 'config/query',
];

/**
 * Host process-control routes (plugin lifecycle, NOT forwarded to gostc).
 * Registered as a separate prefix with no trailing slash (webServer matcher).
 */
export const GOSTC_PROCESS_PREFIX = '/gostc/process/';

/** Document-relative route for a process-control subpath, e.g. gostcProcessRoute('status'). */
export const gostcProcessRoute = (sub) => GOSTC_PROCESS_PREFIX.slice(1) + sub;

/** Allowed process-control subpaths (enforced Host-side). */
export const GOSTC_PROCESS_SUBPATHS = [
  'status', 'start', 'stop', 'restart', 'download',
  'service/install', 'service/uninstall',
];