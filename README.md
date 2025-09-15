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
