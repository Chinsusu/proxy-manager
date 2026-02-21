
# PGW Roadmap

## ✅ v1.0–v1.1 — MVP
- TCP enforcement qua nftables (redirect 80/443 → forwarder :15001)
- Health-check tự động (30s), telemetry: status/latency/exit_ip
- Proxy CRUD, Client CRUD, Mapping CRUD
- JWT authentication (HS256, Argon2id passwords)
- pgw-agent: nftables reconcile tự động
- UI placeholder

## ✅ v1.2–v1.3 — Stability
- File store (`PGW_STORE=file`) — persist qua restart
- Forwarder CONNECT với log ẩn nhạy cảm
- Docker Compose + systemd samples

## ✅ v1.4 — UI Overhaul
- Refactor UI: embedded templates (`html/template`), Vuexy Light theme
- Dashboard với real-time cards, Proxy Status table
- Auto health-check tick trong API process

## ✅ v1.5 — Account Management
- **Email management**: CRUD email accounts, liên kết PayPal
- **PayPal management**: balance, status, verified flag
- **Income tracking**: ghi nhận thu nhập, source/month grouping
- **Reports**: Chart.js monthly bar + source doughnut
- 12 API endpoints mới (`/v1/emails`, `/v1/paypals`, `/v1/income`)

## ✅ v1.6 — SQLite Store
- `PGW_STORE=sqlite` — `modernc.org/sqlite` (pure Go, no CGO)
- WAL mode, PRAGMA foreign_keys, busy_timeout
- `SetProxyTelemetryBatch`: 1 transaction cho 20K proxies (~1.4ms)
- `GetIncomeReport`: SQL `SUM + GROUP BY` aggregation
- `ListMappings`: SQL JOIN thay vì 3 map lookups
- Cascade delete PayPal trong transaction

## 🔄 v1.7 — Planned
- [ ] Multi-node deployment improvements
- [ ] Webhook auto-deploy từ GitHub push
- [ ] Proxy import CSV (bulk)
- [ ] Client CIDR rộng hơn /32 (Phương án B)
- [ ] Prometheus metrics endpoint

## 💡 Future Ideas
- HA pair với VRRP
- OIDC SSO
- Multi-tenant orgs/projects
- UDP TPROXY support
- Per-client bandwidth quotas via `tc`
