# PGW — Coding Standards

Tài liệu này quy định coding conventions cho toàn bộ codebase PGW.  
Phải tuân thủ khi tạo file mới hoặc sửa code hiện tại.

---

## Go

### Formatting
- **Bắt buộc** chạy `gofmt` hoặc `goimports` trước khi commit.
- Tab indent, không dùng spaces.
- Line length mềm ~120 ký tự.

### Naming
```go
// Exported: PascalCase
type ProxyStatus string
func NewSQLite(path string) (Store, error)

// Unexported: camelCase
type sqliteStore struct { ... }
func (s *sqliteStore) migrate() error

// Constants: PascalCase nếu exported, camelCase nếu không
const StatusOK ProxyStatus = "OK"
const defaultPath = "/var/lib/pgw/state.db"

// Acronyms: viết hoa hết hoặc thường hết (không lẫn lộn)
// ✅ userID, httpClient, pgwAPI
// ❌ UserId, HttpClient, PgwApi
```

### Error Handling
```go
// Wrap errors với %w để có stack trace
if err != nil {
    return fmt.Errorf("migrate sqlite: %w", err)
}

// Không dùng panic trong production code (chỉ trong init/test)
// Không bỏ qua lỗi quan trọng với _
res, err := db.Exec(...)
if err != nil { return err }
```

### Package Structure
```
cmd/
  api/main.go      — HTTP handlers + store init + server
  ui/main.go       — Template renderer + proxy routes
  agent/main.go    — nftables reconciler
  fwd/main.go      — TCP forwarder
  health/main.go   — Health check loop
pkg/
  types/types.go   — Shared data models (no logic)
  store/           — Store interface + implementations
  auth/            — JWT + password hashing
  config/          — Env var parsing
  check/           — Proxy health check logic
  httpx/           — HTTP utilities
  logging/         — Structured logging
```

### Store Interface
- Mọi storage backend phải implement `store.Store` đầy đủ.
- Constructor return `(Store, error)`, không panic.
- Batch operations dùng transaction.
- SQL aggregation thay vì loop Go cho reports.

### HTTP Handlers
```go
// Pattern chuẩn cho API handler
func handleFoo(st store.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            // ...
        case http.MethodPost:
            // ...
        default:
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        }
    }
}
```

---

## REST API

### Naming
```
GET    /v1/proxies          — list
POST   /v1/proxies          — create
GET    /v1/proxies/{id}     — get one
PUT    /v1/proxies/{id}     — update
DELETE /v1/proxies/{id}     — delete
POST   /v1/proxies/{id}/check — action (verb sau resource)
```

- **Nouns, not verbs** cho resource path.
- **Plural** cho collections.
- **Lowercase, kebab-case** cho multi-word: `/income-report`

### Response format
```json
// Success list
[{ "id": "...", "host": "..." }]

// Success single
{ "id": "...", "host": "..." }

// Error
{ "error": "message ngắn gọn" }
```

- HTTP 200 cho GET/PUT thành công.
- HTTP 201 cho POST tạo resource.
- HTTP 400 cho request invalid.
- HTTP 401 cho unauthorized.
- HTTP 404 cho not found.
- HTTP 500 cho server error (không expose chi tiết nội bộ).

---

## HTML Templates (Go `html/template`)

### File naming
```
cmd/ui/web/templates/
  base.html          — Layout chung (sidebar + content slot)
  dashboard.html     — Trang dashboard
  proxies.html       — Trang proxy management
  <feature>.html     — Tên = route path (lowercase)
```

### Template blocks
```html
{{define "title"}}Tên Trang{{end}}

{{define "content"}}
<!-- nội dung trang -->
{{end}}

{{define "scripts"}}
<script>/* JS riêng cho trang này */</script>
{{end}}
```

### CSS Classes (Vuexy theme)
- Layout: `layout-wrapper`, `layout-menu`, `layout-page`, `content-wrapper`
- Sidebar: `menu-inner`, `menu-item`, `menu-link`, `menu-icon`, `menu-header`
- Cards: `card`, `card-body`, `card-header`
- Badges: `badge bg-label-{color}`
- Buttons: `btn btn-{variant} btn-sm`

---

## JavaScript

### Class structure
```javascript
class FeatureManager {
    constructor() {
        this.apiBase = '/api/v1/feature';
        this.init();
    }

    init() {
        this.bindEvents();
        this.load();
    }

    async load() { /* fetch + render */ }
    async create(data) { /* POST */ }
    async delete(id) { /* DELETE */ }

    bindEvents() { /* addEventListener */ }
    render(items) { /* DOM updates */ }
}

// Khởi tạo theo route
if (window.location.pathname === '/feature') {
    new FeatureManager();
}
```

### Conventions
- ES6+: `const`, `let`, arrow functions, async/await, template literals.
- Không dùng `var`.
- `camelCase` cho functions và variables.
- Prefix `handle` cho event handlers: `handleSubmit`, `handleDelete`.
- Luôn handle fetch errors — hiển thị toast/alert cho user.

### API calls
```javascript
// Chuẩn fetch pattern
async function apiCall(url, method = 'GET', body = null) {
    const opts = { method, headers: { 'Content-Type': 'application/json' } };
    if (body) opts.body = JSON.stringify(body);
    const res = await fetch(url, opts);
    if (!res.ok) throw new Error(await res.text());
    return res.json();
}
```

---

## Git

### Commit message (Conventional Commits)
```
<type>(<scope>): <mô tả ngắn>

<body tùy chọn>
```

**Types:**
| Type | Khi dùng |
|------|----------|
| `feat` | Tính năng mới |
| `fix` | Sửa bug |
| `docs` | Chỉ thay đổi tài liệu |
| `refactor` | Refactor không thay đổi behavior |
| `chore` | Build, deps, config |
| `style` | CSS/UI thay đổi không ảnh hưởng logic |
| `perf` | Cải thiện hiệu năng |

**Scopes:** `api`, `ui`, `store`, `agent`, `fwd`, `health`, `auth`, `docs`

**Ví dụ:**
```
feat(store): add SQLite backend with WAL mode (v1.6.0)
fix(api): return 404 instead of 500 for missing proxy
docs: update CONFIG_REFERENCE with sqlite env vars
chore: remove stale binary files from root
```

### Branch strategy
- `main` — production-ready, protected
- Mọi thay đổi commit thẳng vào `main` (repo nhỏ, 1 dev)
- Tag version khi release: `git tag v1.6.0`

### Push (tránh treo)
```bash
GIT_SSH_COMMAND="ssh -o ConnectTimeout=10 -o ServerAliveInterval=10 -o ServerAliveCountMax=3" \
timeout 60 git push origin main
```

---

## Makefile Targets

| Target | Chức năng |
|--------|-----------|
| `make build` | Build tất cả 5 binaries vào `bin/` |
| `make api` | Build chỉ `pgw-api` |
| `make ui` | Build chỉ `pgw-ui` |
| `make clean` | Xóa `bin/` |

---

## Secrets & Security

- **Không hardcode** secrets, passwords, tokens vào source code.
- Mọi secret đọc từ env vars (`.env` file hoặc systemd `EnvironmentFile`).
- Password admin phải dùng `PGW_ADMIN_PASS_HASH` (Argon2id) — không dùng `PGW_ADMIN_PASS` trong production.
- `.gitignore` phải exclude: `*.env`, `state.db`, `state.json`, `nodes.txt`, `*.log`.
