# Troubleshooting

## Client không ra Internet

**Kiểm tra trên node (SSH vào node):**

```bash
# 1. ip_forward bật chưa?
sysctl net.ipv4.ip_forward
# Phải là 1. Nếu 0: echo 1 > /proc/sys/net/ipv4/ip_forward

# 2. dnsmasq chạy chưa?
systemctl is-active dnsmasq
ss -tlunp | grep 53   # Phải thấy dnsmasq listen 192.168.2.1:53

# 3. pgw-fwd chạy chưa?
ss -tlnp | grep -E "15001|15002|15003"
systemctl status pgw-fwd@15001

# 4. nftables rules đúng chưa?
nft list table ip pgw
nft list table inet pgw_filter
```

**Client cần:**
- Gateway = `192.168.2.1`
- DNS = `192.168.2.1`
- Chỉ port 80 và 443 được forward (ping bị chặn theo thiết kế).

## pgw-fwd bị 403 Forbidden từ upstream proxy

- Proxy credentials sai hoặc proxy từ chối whitelist IP của node.
- Kiểm tra health proxy: `GET /v1/proxies/{id}/check`.
- Xem log: `journalctl -u pgw-fwd@15001 -n 50 --no-pager`.

## nftables rules không xuất hiện sau khi tạo mapping

```bash
# Trên master: gọi reconcile thủ công
curl -s http://127.0.0.1:9090/agent/reconcile
# Xem log agent
journalctl -u pgw-agent -n 100 --no-pager
```

- Agent fetch `/v1/mappings` — chỉ sinh rule cho trạng thái **APPLIED** hoặc **PENDING**.
- Mapping **FAILED** = proxy health thất bại → không có rule.

## pgw-node không heartbeat / node offline

```bash
# Trên node
journalctl -u pgw-node -n 50 --no-pager
# Kiểm tra kết nối tới master
curl -s https://pgw.eaktur.com/v1/proxies | head -5
```

- Kiểm tra `PGW_API_BASE` trong `/etc/pgw/pgw.env` — phải là URL public của master.
- `PGW_AGENT_TOKEN` phải hợp lệ (JWT chưa hết hạn).

## Region/ISP không hiển thị

- node-side check dùng `ip-api.com` để lookup geo.
- Kiểm tra log pgw-node: `journalctl -u pgw-node -n 50`.
- Đảm bảo node có thể kết nối internet (eth0) để gọi ip-api.com.

## deploy script "Text file busy"

- Không dừng service trước khi ghi binary. Dùng atomic rename:
  ```bash
  # Đúng:
  base64 -d > /usr/local/bin/pgw-node.new && mv /usr/local/bin/pgw-node.new /usr/local/bin/pgw-node
  # Sai: ghi thẳng vào file đang chạy
  ```

## UI trắng / không dữ liệu

- SQLite store dữ liệu trong `/var/lib/pgw/state.db` — persist qua restart.
- Kiểm tra `PGW_STORE=sqlite` và `PGW_STORE_PATH` trong `/etc/pgw/pgw.env`.

## "delete table ... not found" khi reconcile

- Có thể bỏ qua — idempotent. Script xóa table nếu có, không lỗi nếu chưa tồn tại.

## Lệnh nhanh xem trạng thái

```bash
# Master
systemctl status pgw-api pgw-ui pgw-health

# Node
systemctl status pgw-node pgw-agent dnsmasq
journalctl -u pgw-agent -n 50 --no-pager
nft list ruleset
ss -tlnp | grep -E "9090|15001|15002|15003"
sysctl net.ipv4.ip_forward
```
