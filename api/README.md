# API Placeholder

Bạn có thể triển khai API bằng Go/Node/Python. Yêu cầu chính:
- REST endpoints theo docs/API_CONTRACT.md
- JWT auth (HS256) lấy secret từ ENV `API_JWT_SECRET`
- Kết nối Postgres qua `DATABASE_URL`
- Seed admin từ ENV `API_ADMIN_EMAIL` / `API_ADMIN_PASSWORD` nếu chưa tồn tại
- Endpoint cho Agent Pull: `GET /api/v1/agents/:id/pull?since=<version>` trả về 200 hoặc 204
- Health: `GET /api/v1/admin/health`

Gợi ý Go stack:
- Router: chi tiết như chi tiết framework yêu thích
- ORM: GORM / sqlc
- Migration: goose / migrate
- Hash: bcrypt
- JWT: github.com/golang-jwt/jwt/v5

Dockerfile của API nên expose 8082 và đọc `API_BIND`.
