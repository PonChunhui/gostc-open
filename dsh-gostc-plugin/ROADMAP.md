# Roadmap

Planned directions for **dsh-gostc**. Items are proposals, not commitments; order may change.
Feedback and contributions are welcome via [Issues](https://github.com/PonChunhui/dsh-gostc/issues).

## Near term
- [ ] Real screenshots in `screenshots.json` and a short demo GIF on the README.
- [ ] Publish to npm so the market shows download counts and installs prebuilt.
- [ ] i18n polish: review English copy and add more locales if demand appears.
- [ ] Surface backend FRP errors inline on the client / tunnel cards (not only the Logs tab).

## Mid term
- [ ] Import / export of client & tunnel configuration (JSON) for backup and migration.
- [ ] Bulk start / stop and per-group filtering on the management tabs.
- [ ] Optional health watchdog: notify (toast / badge) when a managed backend exits repeatedly.
- [ ] Linux / macOS service integration (systemd / launchd) alongside the Windows service path.

## Longer term
- [ ] Multi-backend profiles: switch between several gostc platforms / local instances.
- [ ] Traffic / connection statistics panel fed by the gostc logger.
- [ ] Guided first-run wizard (platform → key → first tunnel) for new users.

## Done
- [x] Same-origin management panel (clients / tunnels / P2P / logs).
- [x] Auto-provision + managed launch, OS / arch-aware download, reuse of a healthy backend.
- [x] Windows-service install / uninstall (UAC).
- [x] In-panel Settings persisted to `settings.json`.
- [x] About tab: official deep links, self test, copy diagnostics, FAQ, contact channels,
      changelog / roadmap links.