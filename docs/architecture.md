# Kiến trúc hệ thống

## Tổng quan

Hệ thống PGW (Proxy Gateway) gồm hai tier:

```
┌─────────────────────────────────────────────────────────────────┐
│ MASTER SERVER (proxy-server-local)                              │
│                                                                 │
│  pgw-api (:8080)  ─── SQLite ─── pgw-health                   │
│  pgw-ui  (:8081)                                               │
│  pgw-webhook (:9091)                                           │
└─────────────────────┬───────────────────────────────────────────┘
                      │ HTTPS + JWT
┌─────────────────────▼───────────────────────────────────────────┐
│ NODE VPS (ví dụ: 154.37.91.171)                                │
│                                                                 │
│  pgw-node          ← poll master, check proxy, heartbeat       │
│  pgw-agent (:9090) ← reconcile nftables, bật ip_forward        │
│  pgw-fwd@15001     ← transparent TCP proxy → upstream          │
│  pgw-fwd@15002     ← ...                                       │
│  dnsmasq (:53)     ← DNS cho LAN clients                       │
└─────────────────────────────────────────────────────────────────┘
         │ LAN (ens19, 192.168.2.1/24)
    ┌────┴──────────────────────────┐
    │ Clients (192.168.2.101-103)   │
    └───────────────────────────────┘
```

## Thành phần

### Master Server

| Service | Port | Mô tả |
|---------|------|--------|
| `pgw-api` | `:8080` | REST API: CRUD proxy/client/mapping, health-check, telemetry |
| `pgw-ui` | `:8081` | Web UI, reverse proxy `/api/*` → API |
| `pgw-health` | internal | Health-check và geo-lookup định kỳ (30s) |
| `pgw-webhook` | `:9091` | GitHub webhook auto-deploy |

### Node VPS

| Service | Mô tả |
|---------|--------|
| `pgw-node` | Poll master mỗi 15s: lấy assignments, check proxy health, gửi heartbeat |
| `pgw-agent` | Reconcile nftables mỗi 15s, expose `/agent/reconcile` (:9090), bật ip_forward |
| `pgw-fwd@<port>` | Transparent TCP forwarder (1 instance/port), poll master để lấy upstream credentials |
| `dnsmasq` | DNS server cho LAN clients (192.168.2.1:53) |

## Network flow

```
Client (192.168.2.101) → TCP :443 →
  nftables prerouting [REDIRECT] → :15001 →
    pgw-fwd@15001 [CONNECT tunnel] → upstream http proxy:port →
      Internet
```

## nftables sinh ra bởi pgw-agent

```nft
table ip pgw {
  chain prerouting {
    type nat hook prerouting priority dstnat; policy accept;
    iifname "ens19" ip saddr 192.168.2.101/32 tcp dport {80,443} redirect to :15001
    iifname "ens19" ip saddr 192.168.2.102/32 tcp dport {80,443} redirect to :15002
    iifname "ens19" ip saddr 192.168.2.103/32 tcp dport {80,443} redirect to :15003
  }
}

table inet pgw_filter {
  chain forward {
    type filter hook forward priority filter; policy accept;
    ct state established,related accept
    iifname "ens19" oifname "eth0" meta nfproto ipv6 drop  # block IPv6 leak
    ip saddr 192.168.2.101/32 oifname "eth0" drop          # block direct WAN
    ip saddr 192.168.2.101/32 meta l4proto udp drop        # block UDP
    # (repeat cho các client)
  }

  chain input {
    type filter hook input priority filter; policy accept;
    ct state established,related accept
    iifname "ens19" ip saddr 192.168.2.101/32 udp dport 53 accept  # DNS
    iifname "ens19" ip saddr 192.168.2.101/32 tcp dport 53 accept
    iifname "ens19" ip saddr 192.168.2.101/32 tcp dport 15001 accept
    iifname "ens19" tcp dport 15001-15999 drop  # chặn cross-client
  }
}
```

> Interfaces: `ens19` (LAN), `eth0` (WAN) — cấu hình trong `/etc/pgw/pgw.env`.

## Ràng buộc

- **Client IP**: chỉ chấp nhận `/32`. API tự thêm `/32` nếu không có prefix.
- **Mapping model**: 1 client ↔ 1 port (tái sử dụng nếu cùng client, cấp mới nếu chưa có).
- **Port range**: `PGW_FWD_BASE_PORT`..`PGW_FWD_MAX_PORT` (mặc định 15001..15999).
- **Auto-apply**: mapping chỉ active khi upstream proxy health = OK.

## Data store

SQLite (`PGW_STORE=sqlite`) — single file `/var/lib/pgw/state.db`. WAL mode, transaction batching cho telemetry. Không cần PostgreSQL hay NATS.

## Deploy node (one-command)

Từ master, gọi API:
```
POST /v1/nodes/{id}/deploy
```
API sẽ SSH vào node, build và transfer binaries (`pgw-node`, `pgw-agent`, `pgw-fwd`), cài dnsmasq, cấu hình systemd units, enable ip_forward, configure ens19 netplan.