# Hướng dẫn triển khai (Deployment Guide)

## Kiến trúc 2 tier

Xem [architecture.md](./architecture.md) để hiểu toàn cảnh.

- **Master server**: Một máy chủ chạy `pgw-api`, `pgw-ui`, `pgw-health`, `pgw-webhook`.
- **Node VPS**: Một hoặc nhiều VPS chạy `pgw-node`, `pgw-agent`, `pgw-fwd@*`, `dnsmasq`.

---

## A. Cài đặt Master Server

### 1. Chuẩn bị hệ thống (Ubuntu 22.04)

```bash
apt-get update && apt-get install -y curl nftables ca-certificates
```

### 2. Environment file

```bash
mkdir -p /etc/pgw
cat > /etc/pgw/pgw.env << 'EOF'
# Store
PGW_STORE=sqlite
PGW_STORE_PATH=/var/lib/pgw/state.db

# API
PGW_API_ADDR=:8080
PGW_HEALTH_INTERVAL=30s
PGW_STRICT_OUTPUT=true

# JWT (tối thiểu 32 ký tự, random)
JWT_SECRET=<random-string-min-32-chars>

# Admin credentials
PGW_ADMIN_USER=admin
PGW_ADMIN_PASS=<password>          # API sẽ tự hash Argon2id khi start

# UI
PGW_UI_ADDR=:8081
PGW_UI_API=http://127.0.0.1:8080
PGW_UI_AGENT=http://127.0.0.1:9090/agent

# Agent (nội bộ)
PGW_AGENT_TOKEN=<long-lived-jwt-for-agent>
EOF
```

### 3. Build binaries

```bash
cd /opt/proxy-server-local
go build -o /usr/local/bin/pgw-api     ./cmd/api
go build -o /usr/local/bin/pgw-agent   ./cmd/agent
go build -o /usr/local/bin/pgw-ui      ./cmd/ui
go build -o /usr/local/bin/pgw-health  ./cmd/health
go build -o /usr/local/bin/pgw-webhook ./cmd/webhook
go build -o /usr/local/bin/pgw-fwd     ./cmd/fwd
```

### 4. systemd units

Copy các file từ `docs/deploy/systemd/` hoặc tạo thủ công:

```bash
cp docs/deploy/systemd/pgw-api.service     /etc/systemd/system/
cp docs/deploy/systemd/pgw-ui.service      /etc/systemd/system/
cp docs/deploy/systemd/pgw-health.service  /etc/systemd/system/
# Tạo pgw-fwd@.service template (dùng trên cả master lẫn node)
cat > /etc/systemd/system/pgw-fwd@.service << 'EOF'
[Unit]
Description=PGW Forwarder instance on port %i
After=network.target

[Service]
EnvironmentFile=/etc/pgw/pgw.env
Environment=PGW_FWD_ADDR=:%i
ExecStart=/usr/local/bin/pgw-fwd
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now pgw-api pgw-ui pgw-health
```

### 5. Kiểm tra

```bash
curl -s http://localhost:8080/v1/proxies | head -5
# UI tại http://<master-ip>:8081
```

---

## B. Deploy Node VPS (qua UI/API)

> Node VPS được deploy **tự động từ master** — không cần cài thủ công.

### Từ UI

1. Vào **Nodes** → **Add Node**
2. Điền: Host (IP public của VPS), SSH user/password, LAN Subnet (mặc định `192.168.2.1/24`)
3. Click **Deploy** → master sẽ:
   - SSH vào node
   - Cài `nftables`, `dnsmasq`
   - Bật `net.ipv4.ip_forward = 1`
   - Transfer binaries (`pgw-node`, `pgw-agent`, `pgw-fwd`)
   - Cấu hình systemd units và `/etc/pgw/pgw.env` trên node
   - Cấu hình `ens19` (LAN interface) qua netplan

### Từ API

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"<pass>"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['token'])")

curl -s -X POST http://localhost:8080/v1/nodes/<node-id>/deploy \
  -H "Authorization: Bearer $TOKEN"
```

### Services trên node sau deploy

| Service | Status check |
|---------|-------------|
| `pgw-node` | `systemctl is-active pgw-node` |
| `pgw-agent` | `systemctl is-active pgw-agent` |
| `pgw-fwd@<port>` | `systemctl is-active pgw-fwd@15001` |
| `dnsmasq` | `ss -tlunp \| grep 53` |

---

## C. Cấu hình Client

Client (máy Windows/Linux trong LAN) cần:
- **Default gateway**: `192.168.2.1` (IP ens19 của node)
- **DNS**: `192.168.2.1`

Chỉ HTTP (port 80) và HTTPS (port 443) được forward qua proxy. UDP và traffic khác bị chặn.

---

## D. Tạo Proxy → Client → Mapping

```bash
# 1. Tạo proxy
curl -s -X POST http://localhost:8080/v1/proxies \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type":"http","host":"1.2.3.4","port":8080,"username":"user","password":"pass","enabled":true}'

# 2. Tạo client
curl -s -X POST http://localhost:8080/v1/clients \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"ip_cidr":"192.168.2.101"}'

# 3. Tạo mapping
curl -s -X POST http://localhost:8080/v1/mappings \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"<client-id>","proxy_id":"<proxy-id>"}'
```

---

## E. Cập nhật binaries (không redeploy)

```bash
# Build mới
cd /opt/proxy-server-local && go build -o /tmp/pgw-agent-new ./cmd/agent

# Transfer atomic (không "text file busy")
ssh root@<node-ip> 'base64 -d > /usr/local/bin/pgw-agent.new && chmod +x /usr/local/bin/pgw-agent.new && mv /usr/local/bin/pgw-agent.new /usr/local/bin/pgw-agent && systemctl restart pgw-agent'  <<< "$(base64 /tmp/pgw-agent-new)"
```
