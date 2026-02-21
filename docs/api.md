# API Reference

Base: `http://127.0.0.1:8080`

> Tất cả endpoint (trừ `/v1/health` và `/v1/auth/login`) yêu cầu header `Authorization: Bearer <JWT>`.

## Health
- `GET /v1/health` → `"ok"` *(không cần auth)*

## Auth
- `POST /v1/auth/login` body: `{"Username":"admin","Password":"<pass>"}` *(không cần auth)*
  - → `200 {"token":"<JWT>","role":"admin","expires_at":"<RFC3339>"}` (JWT TTL = **12 giờ**)
  - → `401` nếu sai credentials
  - → `429` nếu vượt rate limit (5 lần thất bại / 15 phút / IP)

## Proxies
- `GET /v1/proxies` → `[]Proxy` (sắp xếp theo host, port, id)
- `POST /v1/proxies` body:
  ```json
  {"type":"http","host":"...","port":24639,"username":"...","password":"...","enabled":true}
  ```
  - `type`: `"http"` hoặc `"socks5"`
  - → `201 Proxy`
  - → `409` nếu proxy trùng (cùng host + port + username + password)
- `DELETE /v1/proxies/{id}` → `204 No Content` (cascade xóa mappings liên quan; async cleanup forwarder + nft)
- `POST /v1/proxies/{id}/check` → `{status, latency_ms, exit_ip, err}` và cập nhật telemetry vào store

## Clients
- `GET /v1/clients` → `[]Client`
- `POST /v1/clients` body:
  ```json
  {"ip_cidr":"192.168.2.3/32","enabled":true}
  ```
  Ghi chú: nếu gửi `"192.168.2.3"` sẽ tự chuyển thành `/32`; prefix `<32` sẽ trả `400`.
- `DELETE /v1/clients/{id}` → `204 No Content` (cascade xóa mappings liên quan)

## Mappings
- `GET /v1/mappings` → `[]MappingView` (kèm derived state, sắp xếp theo client IP)
  - **Được dùng bởi Agent** để lấy tất cả mappings để reconcile nftables
- `GET /v1/mappings/active` → `[]MappingView` (filtered theo `PGW_ENFORCE_HEALTH`)
  - **Được dùng bởi Forwarder** để resolve upstream proxy theo `LocalRedirectPort`
- `POST /v1/mappings` body:
  ```json
  {"client_id":"...","proxy_id":"...","node_id":"<optional-node-uuid>"}
  ```
  - `node_id` (tùy chọn): gán proxy này vào node VPS ngay khi tạo
  - API tự gán `local_redirect_port` (1 client = 1 port cố định, dải 15001–15999)
  - Health-check upstream trước khi apply; nếu fail → `state=FAILED`
  - → `409` nếu proxy đã có mapping khác
  - → `201 MappingView`
- `PUT /v1/mappings/{id}/node` body: `{"node_id":"<uuid>"}` hoặc `{"node_id":""}` để unassign
  - Gán hoặc bỏ node VPS khỏi một mapping đã tồn tại
  - → `204 No Content`
  - → `404` nếu mapping không tồn tại
- `POST /v1/mappings/state/{id}` body: `{"state":"APPLIED|PENDING|FAILED","local_redirect_port":15001}`
  - **Chỉ dành cho Agent** gọi sau khi reconcile (role `admin` hoặc `agent`)
  - → `204 No Content`
- `DELETE /v1/mappings/{id}` → `204 No Content` (async: cleanup flag file + stop forwarder + reconcile)

## Nodes
- `GET /v1/nodes` → `[]Node` (admin JWT)
- `POST /v1/nodes` body:
  ```json
  {"label":"VPS SG","ssh_host":"192.168.1.111","ssh_port":22,"ssh_user":"root","ssh_password":"...","public_key":"<Ed25519 hex, tùy chọn>"}
  ```
  - → `201 Node`
- `GET /v1/nodes/{id}` → `Node`
- `PUT /v1/nodes/{id}` → `200 Node` (cập nhật label, ssh creds, public_key)
- `DELETE /v1/nodes/{id}` → `204 No Content`
- `POST /v1/nodes/{id}/deploy` — trigger SSH auto-deploy (async, logs stream vào `deploy_log`)
  - → `202 {"status":"deploying"}`
- `GET /v1/nodes/{id}/deploy/log` → `{"deploy_status":"...","deploy_log":"..."}` (polling)
- `GET /v1/nodes/{id}/assignments` — **Node agent only** (Ed25519 signed): trả danh sách proxy assignments
  - Headers: `X-Node-ID`, `X-Node-TS`, `X-Node-Sig`
  - → `NodeAssignment {node_id, assignments: [{mapping_id, proxy}]}`
- `POST /v1/nodes/{id}/heartbeat` — **Node agent only**: cập nhật `status=online`, `last_seen`, `version`
  - Body: `{"version":"1.0.0"}`
  - → `204 No Content`

## Agent (port :9090)
- `GET /agent/reconcile` — apply nft idempotent, trả `"ok"`
- `POST /agent/reconcile` — giống GET
- `HEAD /agent/reconcile` — chỉ check status (200 OK)
- `GET /agent/health` — liveness check, trả `"ok"`

> Ghi chú: UI reverse proxy `/agent/*` → `http://127.0.0.1:9090/agent` nên có thể gọi qua `http://127.0.0.1:8081/agent/reconcile`.

## Emails
- `GET /v1/emails` → `[]Email`
- `POST /v1/emails` body: `{"address":"foo@gmail.com","provider":"gmail","password":"...","recovery_email":"...","paypal_id":"<optional>","note":"...","status":"active"}`
  - → `201 Email`
- `PUT /v1/emails/{id}` → `200 Email`
- `DELETE /v1/emails/{id}` → `204 No Content`

## PayPals
- `GET /v1/paypals` → `[]PayPal`
- `POST /v1/paypals` body: `{"email":"pp@example.com","owner_name":"...","verified":false,"balance":0,"currency":"USD","status":"active","note":"..."}`
  - → `201 PayPal`
- `PUT /v1/paypals/{id}` → `200 PayPal`
- `DELETE /v1/paypals/{id}` → `204 No Content` (cascade: xóa income liên quan, unlink emails)

## Income
- `GET /v1/income` → `[]Income` (sắp xếp theo `received_at DESC`)
- `POST /v1/income` body: `{"amount":50.0,"currency":"USD","source":"paypal","paypal_id":"<optional>","description":"...","received_at":"2026-02-01T00:00:00Z"}`
  - → `201 Income`
- `DELETE /v1/income/{id}` → `204 No Content`
- `GET /v1/income/report` → `{"total_usd":float,"count":int,"by_source":{"paypal":50,...},"by_month":{"2026-02":50,...}}`
