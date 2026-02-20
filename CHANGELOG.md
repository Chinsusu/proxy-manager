# Changelog

## [1.6.0] - 2026-02-20

### Added
- **SQLite store backend** — new `PGW_STORE=sqlite` option using `modernc.org/sqlite` (pure Go, no CGO binary stays self-contained)
  - File: `pkg/store/sqlite.go` — full `Store` interface implementation
  - 6 tables: `proxies`, `clients`, `mappings`, `emails`, `paypals`, `income`
  - **WAL mode** (`PRAGMA journal_mode=WAL`) — concurrent reads during writes
  - `SetProxyTelemetryBatch` uses single SQL transaction — critical for 20K proxies (benchmark: 100 updates in ~1.4ms)
  - `GetIncomeReport` uses SQL aggregation (`SUM + GROUP BY`) — O(1) vs O(n) Go loop
  - `ListMappings` uses SQL `JOIN` — 1 query vs 3 separate map lookups
  - Cascade delete PayPal (income + email unlink) in one transaction
- Default path: `PGW_STORE_PATH=/var/lib/pgw/state.db`
- Production: updated `/etc/pgw/pgw.env` to use `PGW_STORE=sqlite`

## [1.5.0] - 2026-02-20

### Added
- **Email account management** — CRUD via `GET/POST /v1/emails`, `DELETE /v1/emails/{id}`
- **PayPal account management** — CRUD + balance tracking via `GET/POST /v1/paypals`, `PUT/DELETE /v1/paypals/{id}`
- **Income tracking** — manual income log via `GET/POST /v1/income`, `DELETE /v1/income/{id}`
- **Income reports API** — `GET /v1/income/report` returns `{total_usd, by_source, by_month, count}`
- **UI: Account Management section** — 4 new pages in sidebar: Emails, PayPal, Income, Reports
- **Reports page** — Chart.js bar chart (income by month) + doughnut chart (by source)
- New data models: `Email`, `PayPal`, `Income`, `IncomeReport` in `pkg/types`

## [1.4.0] - 2026-02-20


### Changed
- **UI architecture refactor** — Moved from embedded HTML string constants in `main.go` to separate template files in `cmd/ui/web/templates/` using Go's `html/template` package with `//go:embed`. Binary remains self-contained.
- **New UI theme: Vuexy Light** — Complete redesign from Bootstrap dark theme to Vuexy-inspired light theme:
  - Font: **Public Sans** (Google Fonts)
  - Primary color: **#7367F0** (indigo)
  - Background: **#F8F7FA**, Cards: **#FFFFFF** with shadow
  - Layout: Top navbar → **Vertical sidebar** (260px fixed) with Bootstrap Icons
  - Light-mode badges, stat cards with colored icons, Vuexy card shadows
- **Login page redesign** — Centered card layout, eye toggle for password, spinner on submit
- **Dashboard** — 4 stat cards (Total Proxies, Healthy, Active Mappings, Last Updated) + System Status panel
- **Responsive** — Sidebar hidden on mobile with hamburger toggle + overlay

### Fixed
- Sort arrow unicode (`\u25B2`/`\u25BC`) was rendered as literal text in table headers due to double-backslash in base64-decoded JS; replaced with literal `▲`/`▼` characters

## [1.3.0] - 2026-02-14

### Added
- **Dark Mode OLED UI theme** — Complete UI redesign using AI-generated design system (UI/UX Pro Max skill). Deep black OLED backgrounds (#020617), green CTAs (#22C55E), Fira Code/Sans typography, WCAG AAA accessibility. Cache-busting version bump to `v=1771068456`.

### Fixed
- **UI server template caching bug** — Removed stale file-based templates in `/usr/local/share/pgw/web/` that overrode embedded templates in binary, causing old UI to persist despite code updates.

## [1.2.0] - 2026-02-14

### Fixed
- **Critical: `ListMappings` bug** — PENDING mappings were invisible because `append` was incorrectly nested inside `if m.LastAppliedAt != nil` block in `memoryStore`. New mappings now appear immediately in UI and Agent.

### Security
- **JWT secret validation at startup** — API refuses to start if `PGW_JWT_SECRET` is the insecure default (`dev-change-me`) or shorter than 32 characters. Bypass with `PGW_JWT_STRICT=false` for development.
- **Login rate limiting** — 5 failed login attempts per IP within 15 minutes triggers `429 Too Many Requests`.
- **Request body size limits** — All POST endpoints now enforce 1 MB max body via `http.MaxBytesReader`.

### Added
- **Unit tests** — 25 tests across `pkg/store`, `pkg/auth`, `pkg/config` covering CRUD, cascade deletes, JWT sign/parse, Argon2id hash/verify, and config validation.
- **Forwarder hot-reload** — `pgw-fwd` now polls API every 30s (configurable via `PGW_FWD_POLL_INTERVAL`) and supports `SIGHUP` for immediate upstream re-resolve. No restart needed when proxy changes.
- **Graceful shutdown** — API server handles `SIGTERM`/`SIGINT` with 10s drain timeout.

### Improved
- **Parallel health checks** — `runHealthTick` now uses a worker pool (max 10 concurrent) instead of sequential checks, significantly faster with many proxies.
- **Structured logging** — Added `log/slog` JSON handler alongside legacy loggers for structured, machine-parseable log output. New code can use `logging.L` (slog) or `logging.With()` for context fields. Legacy `logging.Info/Warn/Error` kept for backward compatibility.
- **Startup config validation** — API validates network interfaces (`PGW_WAN_IFACE`, `PGW_LAN_IFACE`), `nft` binary availability, forwarder port range, and data directory at startup, logging warnings for any issues found.

### Fixed (Disk Full Prevention)
- **Batch telemetry writes** — Health check tick now calls `SetProxyTelemetryBatch()` once per tick instead of `SetProxyTelemetry()` per proxy, reducing disk writes from N×`save()` to 1×`save()` per tick cycle.
- **Sampled forwarder logging** — `pgw-fwd` now logs 1 in every 100 successful connections (configurable via `PGW_FWD_LOG_SAMPLE`), reducing log volume by ~99% under heavy traffic. Errors always logged.
- **Journald size limit** — Added `deploy/journald-pgw.conf` drop-in (`SystemMaxUse=500M`, `SystemKeepFree=1G`, rotate weekly) to prevent unbounded journal growth.
