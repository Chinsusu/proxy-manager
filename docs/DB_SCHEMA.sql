-- PGW SQLite Schema (v1.6.0)
-- Used when PGW_STORE=sqlite
-- Auto-applied by NewSQLite() on startup (CREATE TABLE IF NOT EXISTS)

PRAGMA journal_mode=WAL;
PRAGMA busy_timeout=5000;
PRAGMA foreign_keys=ON;

CREATE TABLE IF NOT EXISTS proxies (
    id              TEXT PRIMARY KEY,
    label           TEXT,
    type            TEXT NOT NULL DEFAULT 'http',   -- 'http' | 'socks5'
    host            TEXT NOT NULL,
    port            INTEGER NOT NULL,
    username        TEXT,
    password        TEXT,
    enabled         INTEGER NOT NULL DEFAULT 1,
    status          TEXT NOT NULL DEFAULT 'DOWN',   -- 'OK' | 'DEGRADED' | 'DOWN'
    latency_ms      INTEGER,
    exit_ip         TEXT,
    last_checked_at TEXT                            -- RFC3339Nano UTC
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_proxies_host_port_user
    ON proxies(host, port, COALESCE(username, ''));

CREATE TABLE IF NOT EXISTS clients (
    id      TEXT PRIMARY KEY,
    ip_cidr TEXT NOT NULL,           -- e.g. '192.168.1.10/32'
    note    TEXT,
    enabled INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS mappings (
    id                  TEXT PRIMARY KEY,
    client_id           TEXT NOT NULL REFERENCES clients(id),
    proxy_id            TEXT NOT NULL REFERENCES proxies(id),
    protocol            TEXT,
    local_redirect_port INTEGER DEFAULT 0,   -- assigned port (15001–15999)
    state               TEXT NOT NULL DEFAULT 'PENDING',  -- 'PENDING' | 'APPLIED' | 'FAILED'
    last_applied_at     TEXT                              -- RFC3339Nano UTC
);

CREATE TABLE IF NOT EXISTS emails (
    id             TEXT PRIMARY KEY,
    address        TEXT NOT NULL,
    provider       TEXT DEFAULT 'other',   -- 'gmail' | 'outlook' | 'yahoo' | 'other'
    password       TEXT,
    recovery_email TEXT,
    paypal_id      TEXT REFERENCES paypals(id) ON DELETE SET NULL,
    note           TEXT,
    status         TEXT NOT NULL DEFAULT 'active',  -- 'active' | 'disabled' | 'banned'
    created_at     TEXT NOT NULL                    -- RFC3339Nano UTC
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_emails_address ON emails(address);

CREATE TABLE IF NOT EXISTS paypals (
    id         TEXT PRIMARY KEY,
    email      TEXT NOT NULL,
    owner_name TEXT,
    verified   INTEGER NOT NULL DEFAULT 0,
    balance    REAL NOT NULL DEFAULT 0,
    currency   TEXT NOT NULL DEFAULT 'USD',
    status     TEXT NOT NULL DEFAULT 'active',  -- 'active' | 'limited' | 'suspended'
    note       TEXT,
    created_at TEXT NOT NULL                    -- RFC3339Nano UTC
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_paypals_email ON paypals(email);

CREATE TABLE IF NOT EXISTS income (
    id          TEXT PRIMARY KEY,
    amount      REAL NOT NULL,
    currency    TEXT NOT NULL DEFAULT 'USD',
    source      TEXT,       -- 'paypal' | 'bank' | 'crypto' | etc.
    paypal_id   TEXT REFERENCES paypals(id) ON DELETE SET NULL,
    description TEXT,
    received_at TEXT NOT NULL,  -- RFC3339Nano UTC — used for monthly grouping
    created_at  TEXT NOT NULL   -- RFC3339Nano UTC
);
CREATE INDEX IF NOT EXISTS idx_income_received ON income(received_at DESC);
