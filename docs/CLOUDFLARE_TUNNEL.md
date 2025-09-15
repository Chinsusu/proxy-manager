# Cloudflare Tunnel

## Cài cloudflared
```bash
curl -L https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -o cloudflared
sudo mv cloudflared /usr/local/bin/
sudo chmod +x /usr/local/bin/cloudflared
```

## Login & tạo tunnel
```bash
cloudflared tunnel login
cloudflared tunnel create proxy-manager-ui
```

Lưu `Tunnel ID`, file cred `/root/.cloudflared/<tunnel-id>.json`.

## Cấu hình route
Tạo `/root/.cloudflared/config.yml`:
```yaml
tunnel: <TUNNEL-ID>
credentials-file: /root/.cloudflared/<TUNNEL-ID>.json

ingress:
  - hostname: proxy-manager-ui.xelu.top
    service: http://localhost:8080
  - service: http_status:404
```

## DNS Route
```bash
cloudflared tunnel route dns proxy-manager-ui proxy-manager-ui.xelu.top
```

## Chạy service
```bash
sudo cloudflared service install
sudo systemctl enable cloudflared
sudo systemctl start cloudflared
sudo systemctl status cloudflared
```

## Kiểm tra
- Truy cập https://proxy-manager-ui.xelu.top
- Logs: `sudo journalctl -u cloudflared -f`

## Troubleshooting
- Kiểm tra tunnel status: `cloudflared tunnel info proxy-manager-ui`
- List tunnels: `cloudflared tunnel list`
- Delete tunnel: `cloudflared tunnel delete proxy-manager-ui`
