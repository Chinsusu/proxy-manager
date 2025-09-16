# API Contract (v1)

Base URL: `/api/v1`
Authentication: `Authorization: Bearer <JWT>` (trừ /auth/*)

## Auth
- `POST /auth/login` 
  - Body: `{ "email": "admin@example.com", "password": "password" }`
  - Response: `{ "access_token": "jwt_token", "expires_in": 3600 }`
- `GET /auth/me` → User info

## Servers
- `GET /servers` → Array of servers
- `POST /servers` 
  - Body: `{ "name": "Server 1", "tags": ["prod"], "wan_iface": "eth0", "lan_iface": "eth1" }`
- `GET /servers/:id` → Server detail
- `PATCH /servers/:id` → Update server
- `DELETE /servers/:id` → Delete server

## Proxies
- `GET /servers/:server_id/proxies` → Array of proxies for server
- `POST /servers/:server_id/proxies`
  - Body: `{ "label": "Proxy 1", "type": "http", "host": "1.2.3.4", "port": 8080, "username": "user", "password": "pass" }`
- `GET /proxies/:id` → Proxy detail
- `PATCH /proxies/:id` → Update proxy
- `DELETE /proxies/:id` → Delete proxy

## Mappings
- `GET /servers/:server_id/mappings` → Array of mappings for server
- `POST /servers/:server_id/mappings`
  - Body: `{ "client_cidr": "192.168.1.0/24", "dst_ports": [80,443], "upstream_proxy_id": 1, "enabled": true, "notes": "Web traffic" }`
- `GET /mappings/:id` → Mapping detail
- `PATCH /mappings/:id` → Update mapping
- `DELETE /mappings/:id` → Delete mapping

## Admin
- `GET /admin/health` → `{ "status": "ok", "timestamp": "2024-01-01T00:00:00Z" }`
- `GET /admin/summary` → `{ "servers": 2, "proxies": 5, "mappings": 10, "active_servers": 1 }`

## Agent Pull
- `GET /agents/:agent_id/pull?since=<version>`
  - Headers: `X-Agent-Token: <agent_secret>`
  - Response 200: `{ "version": 123, "proxies": [...], "mappings": [...] }`
  - Response 204: No changes since version
- `POST /agents/:agent_id/ack` (Optional)
  - Body: `{ "version": 123, "status": "applied" }`

## Error Responses
- 400: Bad Request - `{ "error": "Invalid input", "details": {...} }`
- 401: Unauthorized - `{ "error": "Invalid token" }`
- 403: Forbidden - `{ "error": "Insufficient permissions" }`
- 404: Not Found - `{ "error": "Resource not found" }`
- 500: Server Error - `{ "error": "Internal server error" }`

## 6. Groups

### 6.1 List Groups
```http
GET /api/v1/groups
Authorization: Bearer <token>
```

**Response**
```json
200 OK
[
  {
    "id": 1,
    "name": "Production",
    "description": "Production proxy servers",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "proxies": [...]
  }
]
```

### 6.2 Create Group
```http
POST /api/v1/groups
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Development",
  "description": "Development environment proxies"
}
```

**Response**
```json
201 Created
{
  "id": 2,
  "name": "Development",
  "description": "Development environment proxies",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 6.3 Update Group
```http
PUT /api/v1/groups/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Development Updated",
  "description": "Updated description"
}
```

**Response**
```json
200 OK
{
  "id": 2,
  "name": "Development Updated",
  "description": "Updated description",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:01:00Z"
}
```

### 6.4 Delete Group
```http
DELETE /api/v1/groups/{id}
Authorization: Bearer <token>
```

**Response**
```json
200 OK
{
  "message": "Group deleted successfully"
}

400 Bad Request (if group has proxies)
{
  "error": "Cannot delete group with proxies. Move proxies to another group first."
}
```

### 6.5 Move Proxy to Group
```http
PUT /api/v1/proxies/{id}/group
Authorization: Bearer <token>
Content-Type: application/json

{
  "group_id": 5
}
```

**Response**
```json
200 OK
{
  "id": 123,
  "label": "Proxy-1",
  "group_id": 5,
  // ... other proxy fields
}
```

### 6.6 Bulk Move Proxies
```http
PUT /api/v1/proxies/bulk-move
Authorization: Bearer <token>
Content-Type: application/json

{
  "proxy_ids": [1, 2, 3, 4, 5],
  "group_id": 3
}
```

**Response**
```json
200 OK
{
  "message": "Proxies moved successfully",
  "moved_count": 5
}
```
