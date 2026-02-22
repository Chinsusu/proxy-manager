# QUICK OPS — PGW Proxy Gateway

## Kiểm tra nhanh trạng thái

```bash
# Master
systemctl status pgw-api pgw-ui pgw-health

# Node (SSH vào node)
systemctl is-active pgw-node pgw-agent dnsmasq
sysctl net.ipv4.ip_forward           # phải là 1
ss -tlnp | grep -E "15001|15002|15003"  # pgw-fwd listening
ss -tlunp | grep 53                  # dnsmasq listening
nft list table ip pgw                # redirect rules
```

## Tạo Proxy + Client + Mapping (curl)

```bash
# Login
TOKEN=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"<pass>"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['token'])")

# Tạo proxy
PROXY_ID=$(curl -s -X POST http://localhost:8080/v1/proxies \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type":"http","host":"1.2.3.4","port":8080,"username":"u","password":"p","enabled":true}' \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['id'])")

# Tạo client (IP/32)
CLIENT_ID=$(curl -s -X POST http://localhost:8080/v1/clients \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"ip_cidr":"192.168.2.101"}' \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['id'])")

# Tạo mapping
curl -s -X POST http://localhost:8080/v1/mappings \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"client_id\":\"$CLIENT_ID\",\"proxy_id\":\"$PROXY_ID\"}"
```

## Deploy node mới

```bash
# Thêm node qua API
curl -s -X POST http://localhost:8080/v1/nodes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"host":"154.37.91.171","ssh_user":"root","ssh_pass":"<pass>","lan_subnet":"192.168.2.1/24"}'

# Deploy node (install all binaries + config trên node VPS)
curl -s -X POST http://localhost:8080/v1/nodes/<node-id>/deploy \
  -H "Authorization: Bearer $TOKEN"
```

## Reconcile thủ công

```bash
# Trên node — gọi agent để re-apply nftables
curl -s http://127.0.0.1:9090/agent/reconcile
```

## Restart services

```bash
# Master
systemctl restart pgw-api && systemctl status pgw-api

# Node (SSH)
systemctl restart pgw-node pgw-agent
systemctl restart pgw-fwd@15001 pgw-fwd@15002 pgw-fwd@15003
```

## Xem log

```bash
# Master
journalctl -u pgw-api -n 100 --no-pager
journalctl -u pgw-health -n 50 --no-pager

# Node (SSH)
journalctl -u pgw-node -n 50 --no-pager
journalctl -u pgw-agent -n 50 --no-pager
journalctl -u pgw-fwd@15001 -n 50 --no-pager
```

## Kiểm tra proxy health thủ công

```bash
curl -s -X POST http://localhost:8080/v1/proxies/<id>/check \
  -H "Authorization: Bearer $TOKEN"
```

## Check node status qua SSH (từ master)

```bash
JWT_SECRET=$(grep JWT_SECRET /etc/pgw/pgw.env | cut -d= -f2-)
go run /tmp/check_node_svcs.go  # chạy trên master
```

## Cấu hình client (Windows/Linux)

- **Default gateway**: `192.168.2.1`
- **DNS**: `192.168.2.1`
- Chỉ port 80/443 được forward; ping bị chặn theo thiết kế.
