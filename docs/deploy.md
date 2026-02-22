# Triển khai & Vận hành (Operations)

## Systemd units trên Master

### `/etc/systemd/system/pgw-api.service`
```ini
[Unit]
Description=PGW API
After=network.target

[Service]
EnvironmentFile=/etc/pgw/pgw.env
ExecStart=/usr/local/bin/pgw-api
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

### `/etc/systemd/system/pgw-ui.service`
```ini
[Unit]
Description=PGW UI
After=network.target pgw-api.service

[Service]
EnvironmentFile=/etc/pgw/pgw.env
ExecStart=/usr/local/bin/pgw-ui
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

### `/etc/systemd/system/pgw-health.service`
```ini
[Unit]
Description=PGW Health Checker
After=network.target pgw-api.service

[Service]
EnvironmentFile=/etc/pgw/pgw.env
ExecStart=/usr/local/bin/pgw-health
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

## Systemd units trên Node (auto-generated bởi deploy)

### `/etc/systemd/system/pgw-node.service`
```ini
[Unit]
Description=PGW Node Agent
After=network.target

[Service]
EnvironmentFile=/etc/pgw/pgw.env
ExecStart=/usr/local/bin/pgw-node
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

### `/etc/systemd/system/pgw-agent.service`
```ini
[Unit]
Description=PGW Agent (nftables reconcile)
After=network.target

[Service]
EnvironmentFile=/etc/pgw/pgw.env
ExecStart=/usr/local/bin/pgw-agent
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

### `/etc/systemd/system/pgw-fwd@.service`
```ini
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
```

> `pgw-fwd` bind `:%i` (all interfaces) vì nftables `redirect` chỉ chuyển traffic từ client đến localhost — không có routing loop vì traffic OUT từ pgw-fwd ra WAN (eth0) không bị intercept.

## dnsmasq trên Node

Config `/etc/dnsmasq.d/pgw.conf`:
```
interface=ens19
bind-interfaces
no-resolv
server=8.8.8.8
server=1.1.1.1
```

## netplan LAN interface (ens19)

File `/etc/netplan/01-pgw-lan.yaml` (auto-generated):
```yaml
network:
  version: 2
  ethernets:
    ens19:
      dhcp4: false
      addresses: [192.168.2.1/24]
```

## Authentication (JWT)

- `POST /v1/auth/login` với `{username, password}` → trả JWT token
- UI: trang `/login`, cookie `pgw_jwt`
- Agent/Node: dùng `PGW_AGENT_TOKEN` (JWT dài hạn, role `agent`)
- Token admin: `PGW_JWT_SECRET` tối thiểu 32 chars

## Webhook auto-deploy (optional)

Xem [`webhook-setup.md`](./webhook-setup.md) để cấu hình GitHub webhook → tự build và deploy khi push `main`.

```bash
# Test webhook
curl -X POST http://localhost:9091/webhook -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: sha256=..." -d '{"ref":"refs/heads/main"}'
```
