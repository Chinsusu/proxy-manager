# Security Notes

## JWT Security
- JWT secret dài, ngẫu nhiên (>= 32 bytes).
- Algorithm: HS256
- Token expiry: 1 hour (configurable)
- Refresh token: Optional implementation

## Password Security
- Hash password bằng bcrypt (cost >= 12).
- Password strength requirements: min 8 chars, mixed case, numbers
- No password in logs or responses

## API Security
- All endpoints except /auth/* and /admin/health require JWT
- Input validation on all endpoints
- SQL injection protection (parameterized queries)
- Rate limiting on auth endpoints

## Network Security
- Nginx chỉ expose :8080 (UI + /api).
- No direct database access from outside
- Agent authentication via X-Agent-Token header
- HTTPS only in production (via Cloudflare)

## Cloudflare Tunnel Security
- Zero-trust network access
- Optional: Cloudflare Access Policy (email allowlist)
- No direct server IP exposure
- DDoS protection via Cloudflare

## Database Security
- Separate user for application (not postgres superuser)
- Connection via internal Docker network only
- Regular backups with encryption
- Environment variables for credentials (not hardcoded)

## Logging Security
- Không log credentials/secrets
- Log authentication attempts
- Structured logging with levels
- Audit trail cho admin actions

## Best Practices
- Keep Docker images updated
- Regular security patches
- Monitor unusual access patterns
- Principle of least privilege
- Secrets rotation policy
