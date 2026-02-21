# PGW — Proxy Gateway Manager

PGW ép toàn bộ HTTP/HTTPS (TCP 80/443) từ **client trong LAN** đi qua **forwarder (:15001)**, sử dụng **upstream HTTP/SOCKS5 proxy** để đổi **exit IP**.

**Phiên bản hiện tại: v1.6.0**

---

## Kiến trúc

| Component | Binary | Port | Chức năng |
|-----------|--------|------|-----------|
| API | `pgw-api` | 8080 | REST API: proxies/clients/mappings/email/paypal/income, health-check, telemetry |
| Agent | `pgw-agent` | 9090 | Sinh & apply rules **nftables** từ mappings |
| Forwarder | `pgw-fwd` | 15001 | Transparent CONNECT, log ẩn nhạy cảm |
| UI | `pgw-ui` | 8081 | Dashboard + reverse proxy (`/api/*` → API) |

---

## Quick Start

**Yêu cầu:** Go ≥ 1.23, Linux + nftables, systemd.

```bash
# Build
cd /opt/proxy-server-local
make build

# Install
sudo install -m 0755 bin/pgw-* /usr/local/bin/

# Start
sudo systemctl restart pgw-api pgw-agent pgw-fwd pgw-ui
```

Truy cập UI: `http://<server>:8081`

---

## Cấu hình (Environment Variables)

### API (`/etc/pgw/pgw.env`)

| Biến | Mặc định | Mô tả |
|------|----------|-------|
| `PGW_API_ADDR` | `:8080` | Listen address |
| `PGW_STORE` | `memory` | Storage: `memory`, `file`, `sqlite` |
| `PGW_STORE_PATH` | `/var/lib/pgw/state.db` | Path (cho `file`/`sqlite`) |
| `PGW_HEALTH_INTERVAL` | `30s` | Chu kỳ health-check proxy |
| `PGW_JWT_SECRET` | — | **Bắt buộc** — khoá ký JWT |
| `PGW_ADMIN_USER` | `admin` | Tên đăng nhập |
| `PGW_ADMIN_PASS_HASH` | — | Argon2id PHC hash (khuyến nghị) |
| `PGW_ADMIN_PASS` | — | Plain-text password (không khuyến nghị) |
| `PGW_AGENT_TOKEN` | — | Token nội bộ cho Agent role |

### Agent

| Biến | Mặc định | Mô tả |
|------|----------|-------|
| `PGW_AGENT_ADDR` | `:9090` | Listen address |
| `PGW_API_BASE` | `http://127.0.0.1:8080` | URL của pgw-api |
| `PGW_WAN_IFACE` | — | WAN interface (e.g. `eth0`) |
| `PGW_LAN_IFACE` | — | LAN interface (e.g. `ens19`) |

### Forwarder & UI

| Biến | Mặc định | Mô tả |
|------|----------|-------|
| `PGW_FWD_ADDR` | `:15001` | Forwarder listen |
| `PGW_UI_ADDR` | `:8081` | UI listen |
| `PGW_UI_API` | `http://127.0.0.1:8080` | API URL cho UI |
| `PGW_UI_AGENT` | `http://127.0.0.1:9090/agent` | Agent URL cho UI |

> **Khuyến nghị production:** `PGW_STORE=sqlite`, `PGW_STORE_PATH=/var/lib/pgw/state.db`

---

## Authentication (JWT)

```bash
# Đăng nhập
curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"Username":"admin","Password":"your-password"}' | jq .token

# Sử dụng token
curl -H "Authorization: Bearer <JWT>" http://localhost:8080/v1/proxies
```

Các endpoint trừ `/v1/health` và `/v1/auth/login` đều yêu cầu JWT.

---

## API nhanh

```bash
API=http://127.0.0.1:8080
HDR="-H 'Authorization: Bearer $TOKEN' -H 'Content-Type: application/json'"

# Thêm proxy
curl -s $HDR -X POST $API/v1/proxies \
  -d '{"type":"http","host":"1.2.3.4","port":8080,"username":"user","password":"pass","enabled":true}'

# Thêm client
curl -s $HDR -X POST $API/v1/clients \
  -d '{"ip_cidr":"192.168.1.10/32","enabled":true}'

# Tạo mapping
curl -s $HDR -X POST $API/v1/mappings \
  -d '{"client_id":"<CID>","proxy_id":"<PID>"}'

# Reconcile (apply nftables)
curl -s -X POST http://localhost:9090/agent/reconcile
```

Chi tiết: `docs/api.md` | OpenAPI: `docs/API_SPEC.yaml`

---

## Luồng hoạt động

```
Client (192.168.x.x) → TCP 80/443
  → nftables NAT redirect → :15001 (pgw-fwd)
    → pgw-fwd CONNECT → Upstream Proxy (HTTP/SOCKS5)
      → Target server (exit IP đổi)
```

---

## Tính năng

### Core (v1.0–v1.4)
- Quản lý proxy upstream (HTTP/SOCKS5), client, mapping
- Health-check tự động với telemetry (status/latency/exit_ip)
- nftables rules tự động qua pgw-agent
- JWT authentication
- Dashboard UI (Vuexy theme)

### Account Management (v1.5)
- **Email management**: CRUD email accounts, liên kết PayPal
- **PayPal management**: Quản lý tài khoản PayPal, balance
- **Income tracking**: Ghi nhận thu nhập, nhóm theo nguồn/tháng
- **Reports**: Biểu đồ Chart.js (monthly bar, source doughnut)

### SQLite Store (v1.6)
- Store backend: `PGW_STORE=sqlite` — dùng `modernc.org/sqlite` (pure Go)
- WAL mode: concurrent reads khi write
- Batch telemetry: 20K proxy updates trong 1 transaction (~1.4ms)
- SQL aggregation cho income reports

---

## Update

### Tự động (khuyến nghị)
```bash
bash /opt/proxy-server-local/update-pgw.sh
```

### Thủ công
```bash
cd /opt/proxy-server-local
git pull origin main
make build
sudo systemctl stop pgw-api pgw-ui pgw-health pgw-agent
sudo cp bin/pgw-* /usr/local/bin/
sudo systemctl start pgw-agent pgw-api pgw-ui pgw-health
```

### Multi-node
```bash
./update-all-nodes.sh [nodes.txt]
```

---

## Tài liệu

| Tài liệu | Nội dung |
|----------|----------|
| `docs/architecture.md` | Kiến trúc, nftables rules |
| `docs/deploy.md` | Cài đặt, systemd, cấu hình |
| `docs/api.md` | API endpoints |
| `docs/API_SPEC.yaml` | OpenAPI spec |
| `docs/CONFIG_REFERENCE.md` | Tham chiếu đầy đủ env vars |
| `docs/QUICK_OPS.md` | Lệnh vận hành nhanh |
| `docs/troubleshooting.md` | Xử lý lỗi |
| `docs/security.md` | Bảo mật |
| `docs/CODING_STANDARDS.md` | Coding standard |
| `CHANGELOG.md` | Lịch sử thay đổi |

---

## Giới hạn

- Client chỉ hỗ trợ `/32` (single IP)
- Upstream proxy: `http` và `socks5`
- `memory` store mất data khi restart → dùng `sqlite`

---

## License

AGPL-3.0
