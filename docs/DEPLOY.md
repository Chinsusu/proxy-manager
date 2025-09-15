# Triển khai Production (1 server)

## Yêu cầu
- Ubuntu 22.04+ hoặc tương đương
- Docker + Docker Compose
- Domain: proxy-manager-ui.xelu.top (Cloudflare)
- Cloudflare Tunnel (cloudflared)

## Bước 1: Clone & ENV
```bash
git clone <repo>
cd proxy-manager-control-plane
cp .env.example .env

# sửa API_ADMIN_EMAIL / API_ADMIN_PASSWORD / API_JWT_SECRET
./scripts/gen_jwt_secret.sh  # copy output vào .env
```

## Bước 2: Up
```bash
make up
```
Kiểm tra:
- Nginx local: http://SERVER_IP:8080
- API health: http://SERVER_IP:8080/api/v1/admin/health

## Bước 3: Cloudflare Tunnel
Xem `docs/CLOUDFLARE_TUNNEL.md` để map `proxy-manager-ui.xelu.top` → `http://127.0.0.1:8080`.

## Bước 4: Bảo mật
- Bật Cloudflare Access (nếu muốn).
- Đổi JWT secret, mật khẩu admin mạnh.
- Theo dõi logs: `make logs`.

## Bước 5: Backup
```bash
# Backup database
docker compose exec db pg_dump -U pgm pgm_db > backup.sql

# Restore
docker compose exec -T db psql -U pgm pgm_db < backup.sql
```
