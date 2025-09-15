# Kiến trúc

- Nginx (:8080) phục vụ UI (SPA) + reverse proxy /api/* → API (:8082).
- API (Manager) kết nối Postgres, phát hành JWT, CRUD Servers/Proxies/Mappings, audit.
- Agent (ở các server local) tự `pull` config theo version.
- Cloudflare Tunnel expose `localhost:8080` → `https://proxy-manager-ui.xelu.top`.

## Luồng thay đổi
UI → API (update DB) → tăng `config_version(server_id)` → Agent poll `/agents/{id}/pull?since=<ver>` → nhận config mới → áp dụng → (tuỳ chọn) ack.

## Data Flow
```
[UI Browser] → [Nginx :8080] → [API :8082] → [Postgres]
                     ↓
[Agent Local] ← [API :8082/agents/{id}/pull]
```

## Components
- **UI**: React/Vue SPA build to static files
- **API**: REST API server (Go/Node.js/Python)
- **Database**: PostgreSQL with versioned config
- **Nginx**: Static file server + reverse proxy
- **Agent**: Lightweight client polls config changes
