# Proxy Manager Control-Plane (no Cloudflare Workers)

UI + API quản trị nhiều proxy server local. Truy cập qua Cloudflare Tunnel tại `https://proxy-manager-ui.xelu.top`.

- UI: SPA (build → static) phục vụ qua Nginx.
- API: Manager API (REST) + Postgres.
- Agent (ở local) **pull** config định kỳ từ API.

## Tính năng
- Quản lý Servers (node/agent), Proxies, Mappings.
- Chỉ 1 user: admin (seed từ ENV).
- Health / Summary.
- Versioned Config cho Agent Pull.

## Quickstart
```bash
cp .env.example .env
# chỉnh các biến trong .env
make up
```

UI: [http://localhost:8080](http://localhost:8080)
API: [http://localhost:8080/api/v1](http://localhost:8080/api/v1)

Xem `docs/DEPLOY.md` để chạy production (Nginx + Cloudflare Tunnel).

## Next Steps

### 1. Implement API
- Vào `api/` directory và implement REST API theo `docs/API_CONTRACT.md`
- Build Docker image và push lên registry (ghcr.io/your-org/pgm-api:latest)
- Update `docker-compose.yml` với image thực tế

### 2. Implement UI  
- Vào `ui/` directory và tạo SPA theo yêu cầu
- Build static files và setup cách copy vào `ui_build` volume
- Update `docker-compose.yml` với UI image thực tế

### 3. Deploy
- Copy repo lên server production
- Chạy `make up` để test local
- Setup Cloudflare Tunnel theo `docs/CLOUDFLARE_TUNNEL.md`
- Access via https://proxy-manager-ui.xelu.top

### 4. Agent Development
- Tạo agent client (Go/Python/Rust) gọi `/agents/{id}/pull` API
- Test với local proxy server

Repo này chứa toàn bộ infrastructure code và documentation để bạn tập trung vào business logic!
