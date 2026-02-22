# PGW Roadmap

## ✅ v1.0–v1.1 — MVP
- TCP enforcement qua nftables (redirect 80/443 → forwarder :15001)
- Health-check tự động (30s), telemetry: status/latency/exit_ip
- Proxy CRUD, Client CRUD, Mapping CRUD
- JWT authentication (HS256, Argon2id passwords)
- pgw-agent: nftables reconcile tự động

## ✅ v1.2–v1.3 — Stability
- File store (`PGW_STORE=file`) — persist qua restart
- Forwarder CONNECT với log ẩn nhạy cảm
- Systemd bare-metal deployment

## ✅ v1.4 — UI Overhaul
- Web UI: embedded templates, dashboard, proxy status table
- Auto health-check trong API process

## ✅ v1.5 — Account Management
- Email CRUD, PayPal management, income tracking
- Reports: Chart.js monthly/source grouping

## ✅ v1.6 — SQLite Store
- `PGW_STORE=sqlite` (pure Go, no CGO)
- WAL mode, batch telemetry, SQL aggregation
- ListMappings via SQL JOIN

## ✅ v1.7 — Multi-Node Deployment
- Node management: CRUD, deploy via SSH
- `pgw-node`: poll master, check proxy health, heartbeat
- One-click deploy từ UI: transfer binaries, config systemd, netplan ens19
- `pgw-agent`: enable ip_forward khi khởi động
- `dnsmasq`: cài tự động trên node để phục vụ DNS cho LAN clients
- SOCKS5 upstream proxy support trong pgw-fwd
- Geo-lookup region/ISP từ node-side (ip-api.com)
- LAN subnet (`lan_subnet`) configurable per node

## 🔄 v1.8 — Planned
- [ ] Webhook auto-deploy từ GitHub push (đã có pgw-webhook, chưa tích hợp full)
- [ ] Proxy import CSV (bulk)
- [ ] Prometheus metrics endpoint `/metrics`
- [ ] Node health dashboard — show all services status per node

## 💡 Future Ideas
- HA pair với VRRP
- OIDC SSO
- Multi-tenant orgs/projects
- UDP TPROXY support
- Per-client bandwidth quotas via `tc`
- IPv6 support (hiện chặn toàn bộ IPv6 để tránh leak)
