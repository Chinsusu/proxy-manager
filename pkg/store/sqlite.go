// pkg/store/sqlite.go
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"

	"github.com/Chinsusu/proxy-server-local/pkg/types"
)

// sqliteStore implements Store using SQLite via modernc.org/sqlite (pure Go, no CGO).
type sqliteStore struct {
	db *sql.DB
}

// NewSQLite opens (or creates) a SQLite database at path and runs auto-migration.
func NewSQLite(path string) (Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// Single writer — SQLite handles this best with 1 connection for writes.
	// WAL allows concurrent reads during writes.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		return nil, fmt.Errorf("WAL pragma: %w", err)
	}
	if _, err := db.Exec(`PRAGMA busy_timeout=5000`); err != nil {
		return nil, fmt.Errorf("busy_timeout pragma: %w", err)
	}
	if _, err := db.Exec(`PRAGMA foreign_keys=ON`); err != nil {
		return nil, fmt.Errorf("foreign_keys pragma: %w", err)
	}
	s := &sqliteStore{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *sqliteStore) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS proxies (
			id TEXT PRIMARY KEY,
			label TEXT, type TEXT NOT NULL DEFAULT 'http',
			host TEXT NOT NULL, port INTEGER NOT NULL,
			username TEXT, password TEXT,
			enabled INTEGER NOT NULL DEFAULT 1,
			status TEXT NOT NULL DEFAULT 'DOWN',
			latency_ms INTEGER, exit_ip TEXT, last_checked_at TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_proxies_host_port_user ON proxies(host, port, COALESCE(username,''))`,
		`CREATE TABLE IF NOT EXISTS clients (
			id TEXT PRIMARY KEY,
			ip_cidr TEXT NOT NULL,
			note TEXT,
			enabled INTEGER NOT NULL DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS mappings (
			id TEXT PRIMARY KEY,
			client_id TEXT NOT NULL,
			proxy_id TEXT NOT NULL,
			protocol TEXT,
			local_redirect_port INTEGER DEFAULT 0,
			state TEXT NOT NULL DEFAULT 'PENDING',
			last_applied_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS emails (
			id TEXT PRIMARY KEY,
			address TEXT NOT NULL,
			provider TEXT DEFAULT 'other',
			password TEXT,
			recovery_email TEXT,
			paypal_id TEXT,
			note TEXT,
			status TEXT NOT NULL DEFAULT 'active',
			created_at TEXT NOT NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_emails_address ON emails(address)`,
		`CREATE TABLE IF NOT EXISTS paypals (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL,
			owner_name TEXT,
			verified INTEGER NOT NULL DEFAULT 0,
			balance REAL NOT NULL DEFAULT 0,
			currency TEXT NOT NULL DEFAULT 'USD',
			status TEXT NOT NULL DEFAULT 'active',
			note TEXT,
			created_at TEXT NOT NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_paypals_email ON paypals(email)`,
		`CREATE TABLE IF NOT EXISTS income (
			id TEXT PRIMARY KEY,
			amount REAL NOT NULL,
			currency TEXT NOT NULL DEFAULT 'USD',
			source TEXT,
			paypal_id TEXT,
			description TEXT,
			received_at TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_income_received ON income(received_at DESC)`,
		`CREATE TABLE IF NOT EXISTS nodes (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			public_key TEXT,
			ssh_host TEXT,
			ssh_port INTEGER NOT NULL DEFAULT 22,
			ssh_user TEXT NOT NULL DEFAULT 'root',
			ssh_password TEXT,
			ssh_key TEXT,
			status TEXT NOT NULL DEFAULT 'offline',
			version TEXT,
			last_seen_at TEXT,
			deploy_status TEXT NOT NULL DEFAULT 'pending',
			deploy_log TEXT,
			deployed_at TEXT,
			created_at TEXT NOT NULL
		)`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("migrate stmt %q: %w", stmt[:min(40, len(stmt))], err)
		}
	}
	// Additive column migrations (idempotent — ignore 'duplicate column' errors)
	for _, alter := range []string{
		`ALTER TABLE proxies ADD COLUMN node_id TEXT`,
		`ALTER TABLE mappings ADD COLUMN node_id TEXT`,
		`ALTER TABLE proxies ADD COLUMN region TEXT`,
		`ALTER TABLE proxies ADD COLUMN isp TEXT`,
		`ALTER TABLE nodes ADD COLUMN email_id TEXT`,
		`ALTER TABLE nodes ADD COLUMN lan_subnet TEXT`,
		`ALTER TABLE nodes ADD COLUMN enabled INTEGER NOT NULL DEFAULT 1`,
	} {
		_, _ = s.db.Exec(alter) // intentionally ignore error (column may already exist)
	}
	return nil
}

func min(a, b int) int {
	if a < b { return a }
	return b
}

// ---------- Helpers ----------

func nullStr(s *string) interface{} {
	if s == nil { return nil }
	return *s
}

func nullInt(i *int) interface{} {
	if i == nil { return nil }
	return *i
}

func parseTime(s string) *time.Time {
	if s == "" { return nil }
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil { return nil }
	return &t
}

func fmtTime(t *time.Time) string {
	if t == nil { return "" }
	return t.UTC().Format(time.RFC3339Nano)
}

// ---------- Proxies ----------

func (s *sqliteStore) ListProxies() []types.Proxy {
	rows, err := s.db.Query(`SELECT id,label,type,host,port,username,password,enabled,status,latency_ms,exit_ip,last_checked_at,node_id,region,isp FROM proxies ORDER BY host,port,id`)
	if err != nil { return nil }
	defer rows.Close()
	var out []types.Proxy
	for rows.Next() {
		var p types.Proxy
		var label, username, password, exitIP, lastChecked, nodeID, region, isp sql.NullString
		var latencyMs sql.NullInt64
		_ = rows.Scan(&p.ID, &label, &p.Type, &p.Host, &p.Port, &username, &password, &p.Enabled, &p.Status, &latencyMs, &exitIP, &lastChecked, &nodeID, &region, &isp)
		if label.Valid && label.String != "" { p.Label = label.String }
		if username.Valid && username.String != "" { p.Username = &username.String }
		if password.Valid && password.String != "" { p.Password = &password.String }
		if latencyMs.Valid { v := int(latencyMs.Int64); p.LatencyMs = &v }
		if exitIP.Valid && exitIP.String != "" { p.ExitIP = &exitIP.String }
		if lastChecked.Valid { p.LastCheckedAt = parseTime(lastChecked.String) }
		if nodeID.Valid && nodeID.String != "" { p.NodeID = &nodeID.String }
		if region.Valid && region.String != "" { p.Region = &region.String }
		if isp.Valid && isp.String != "" { p.ISP = &isp.String }
		out = append(out, p)
	}
	return out
}

func (s *sqliteStore) CreateProxy(p types.Proxy) types.Proxy {
	if p.ID == "" { p.ID = uuid.New().String() }
	p.Status = types.StatusDown
	_, _ = s.db.Exec(
		`INSERT OR IGNORE INTO proxies(id,label,type,host,port,username,password,enabled,status) VALUES(?,?,?,?,?,?,?,?,?)`,
		p.ID, p.Label, p.Type, p.Host, p.Port, nullStr(p.Username), nullStr(p.Password), p.Enabled, string(p.Status),
	)
	return p
}

func (s *sqliteStore) UpdateProxy(p types.Proxy) (types.Proxy, bool) {
	res, err := s.db.Exec(
		`UPDATE proxies SET label=?,type=?,host=?,port=?,username=?,password=?,enabled=?,status=? WHERE id=?`,
		p.Label, p.Type, p.Host, p.Port, nullStr(p.Username), nullStr(p.Password), p.Enabled, string(p.Status), p.ID,
	)
	if err != nil { return types.Proxy{}, false }
	n, _ := res.RowsAffected()
	return p, n > 0
}

func (s *sqliteStore) UpdateProxyNode(id, nodeID string) bool {
	var err error
	if nodeID == "" {
		_, err = s.db.Exec(`UPDATE proxies SET node_id=NULL WHERE id=?`, id)
	} else {
		_, err = s.db.Exec(`UPDATE proxies SET node_id=? WHERE id=?`, nodeID, id)
	}
	return err == nil
}

func (s *sqliteStore) DeleteProxy(id string) bool {
	res, _ := s.db.Exec(`DELETE FROM proxies WHERE id=?`, id)
	n, _ := res.RowsAffected()
	if n > 0 { _, _ = s.db.Exec(`DELETE FROM mappings WHERE proxy_id=?`, id) }
	return n > 0
}

// ---------- Clients ----------

func (s *sqliteStore) ListClients() []types.Client {
	rows, err := s.db.Query(`SELECT id,ip_cidr,note,enabled FROM clients ORDER BY ip_cidr`)
	if err != nil { return nil }
	defer rows.Close()
	var out []types.Client
	for rows.Next() {
		var c types.Client
		var note sql.NullString
		_ = rows.Scan(&c.ID, &c.IPCidr, &note, &c.Enabled)
		if note.Valid { c.Note = note.String }
		out = append(out, c)
	}
	return out
}

func (s *sqliteStore) CreateClient(c types.Client) types.Client {
	if c.ID == "" { c.ID = uuid.New().String() }
	_, _ = s.db.Exec(`INSERT OR IGNORE INTO clients(id,ip_cidr,note,enabled) VALUES(?,?,?,?)`, c.ID, c.IPCidr, c.Note, c.Enabled)
	return c
}

func (s *sqliteStore) DeleteClient(id string) bool {
	res, _ := s.db.Exec(`DELETE FROM clients WHERE id=?`, id)
	n, _ := res.RowsAffected()
	if n > 0 { _, _ = s.db.Exec(`DELETE FROM mappings WHERE client_id=?`, id) }
	return n > 0
}

// ---------- Mappings ----------

func (s *sqliteStore) ListMappings() []types.MappingView {
	rows, err := s.db.Query(`
		SELECT m.id, m.state, m.local_redirect_port, m.last_applied_at,
		       c.id, c.ip_cidr, c.note, c.enabled,
		       p.id, p.label, p.type, p.host, p.port, p.username, p.password, p.enabled, p.status, p.latency_ms, p.exit_ip, p.last_checked_at
		FROM mappings m
		JOIN clients c ON c.id = m.client_id
		JOIN proxies p ON p.id = m.proxy_id
		ORDER BY m.last_applied_at DESC NULLS LAST, m.id ASC`)
	if err != nil { return nil }
	defer rows.Close()
	var out []types.MappingView
	for rows.Next() {
		var mv types.MappingView
		var lastApplied, cNote, pLabel, pUsername, pPassword, pExitIP, pLastChecked sql.NullString
		var pLatencyMs sql.NullInt64
		_ = rows.Scan(
			&mv.ID, &mv.State, &mv.LocalRedirectPort, &lastApplied,
			&mv.Client.ID, &mv.Client.IPCidr, &cNote, &mv.Client.Enabled,
			&mv.Proxy.ID, &pLabel, &mv.Proxy.Type, &mv.Proxy.Host, &mv.Proxy.Port,
			&pUsername, &pPassword, &mv.Proxy.Enabled, &mv.Proxy.Status,
			&pLatencyMs, &pExitIP, &pLastChecked,
		)
		if cNote.Valid { mv.Client.Note = cNote.String }
		if pLabel.Valid { mv.Proxy.Label = pLabel.String }
		if pUsername.Valid && pUsername.String != "" { mv.Proxy.Username = &pUsername.String }
		if pPassword.Valid && pPassword.String != "" { mv.Proxy.Password = &pPassword.String }
		if pLatencyMs.Valid { v := int(pLatencyMs.Int64); mv.Proxy.LatencyMs = &v }
		if pExitIP.Valid && pExitIP.String != "" { mv.Proxy.ExitIP = &pExitIP.String }
		if pLastChecked.Valid { mv.Proxy.LastCheckedAt = parseTime(pLastChecked.String) }
		out = append(out, mv)
	}
	return out
}

func (s *sqliteStore) CreateMapping(m types.Mapping) (types.MappingView, bool) {
	if m.ID == "" { m.ID = uuid.New().String() }
	m.State = "PENDING"
	_, err := s.db.Exec(
		`INSERT INTO mappings(id,client_id,proxy_id,protocol,local_redirect_port,state) VALUES(?,?,?,?,?,?)`,
		m.ID, m.ClientID, m.ProxyID, m.Protocol, m.LocalRedirectPort, m.State,
	)
	if err != nil { return types.MappingView{}, false }
	// Fetch the view
	mvs := s.listMappingByID(m.ID)
	if len(mvs) == 0 { return types.MappingView{}, false }
	return mvs[0], true
}

func (s *sqliteStore) listMappingByID(id string) []types.MappingView {
	rows, err := s.db.Query(`
		SELECT m.id, m.state, m.local_redirect_port, m.last_applied_at,
		       c.id, c.ip_cidr, c.note, c.enabled,
		       p.id, p.label, p.type, p.host, p.port, p.username, p.password, p.enabled, p.status, p.latency_ms, p.exit_ip, p.last_checked_at
		FROM mappings m
		JOIN clients c ON c.id = m.client_id
		JOIN proxies p ON p.id = m.proxy_id
		WHERE m.id=?`, id)
	if err != nil { return nil }
	defer rows.Close()
	var out []types.MappingView
	for rows.Next() {
		var mv types.MappingView
		var lastApplied, cNote, pLabel, pUsername, pPassword, pExitIP, pLastChecked sql.NullString
		var pLatencyMs sql.NullInt64
		_ = rows.Scan(
			&mv.ID, &mv.State, &mv.LocalRedirectPort, &lastApplied,
			&mv.Client.ID, &mv.Client.IPCidr, &cNote, &mv.Client.Enabled,
			&mv.Proxy.ID, &pLabel, &mv.Proxy.Type, &mv.Proxy.Host, &mv.Proxy.Port,
			&pUsername, &pPassword, &mv.Proxy.Enabled, &mv.Proxy.Status,
			&pLatencyMs, &pExitIP, &pLastChecked,
		)
		if cNote.Valid { mv.Client.Note = cNote.String }
		if pLabel.Valid { mv.Proxy.Label = pLabel.String }
		if pUsername.Valid && pUsername.String != "" { mv.Proxy.Username = &pUsername.String }
		if pPassword.Valid && pPassword.String != "" { mv.Proxy.Password = &pPassword.String }
		if pLatencyMs.Valid { v := int(pLatencyMs.Int64); mv.Proxy.LatencyMs = &v }
		if pExitIP.Valid && pExitIP.String != "" { mv.Proxy.ExitIP = &pExitIP.String }
		if pLastChecked.Valid { mv.Proxy.LastCheckedAt = parseTime(pLastChecked.String) }
		out = append(out, mv)
	}
	return out
}

func (s *sqliteStore) DeleteMapping(id string) bool {
	res, _ := s.db.Exec(`DELETE FROM mappings WHERE id=?`, id)
	n, _ := res.RowsAffected()
	return n > 0
}

func (s *sqliteStore) UpdateMappingNode(id, nodeID string) bool {
	var res sql.Result
	var err error
	if nodeID == "" {
		res, err = s.db.Exec(`UPDATE mappings SET node_id=NULL WHERE id=?`, id)
	} else {
		res, err = s.db.Exec(`UPDATE mappings SET node_id=? WHERE id=?`, nodeID, id)
	}
	if err != nil { return false }
	n, _ := res.RowsAffected()
	return n > 0
}

func (s *sqliteStore) UpdateMappingState(id string, state string, localPort int) bool {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	var res sql.Result
	var err error
	if state == "APPLIED" || state == "FAILED" {
		res, err = s.db.Exec(`UPDATE mappings SET state=?, local_redirect_port=CASE WHEN ?>0 THEN ? ELSE local_redirect_port END, last_applied_at=? WHERE id=?`,
			state, localPort, localPort, now, id)
	} else {
		res, err = s.db.Exec(`UPDATE mappings SET state=?, local_redirect_port=CASE WHEN ?>0 THEN ? ELSE local_redirect_port END WHERE id=?`,
			state, localPort, localPort, id)
	}
	if err != nil { return false }
	n, _ := res.RowsAffected()
	return n > 0
}

// ---------- Telemetry ----------

func (s *sqliteStore) SetProxyTelemetry(id string, status types.ProxyStatus, latency int, exitIP, region, isp string) {
	s.SetProxyTelemetryBatch([]TelemetryUpdate{{ID: id, Status: status, Latency: latency, ExitIP: exitIP, Region: region, ISP: isp}})
}

// SetProxyTelemetryBatch uses a single transaction for all updates — critical for 20K proxies.
func (s *sqliteStore) SetProxyTelemetryBatch(updates []TelemetryUpdate) {
	if len(updates) == 0 { return }
	tx, err := s.db.Begin()
	if err != nil { return }
	defer func() {
		if err != nil { _ = tx.Rollback() }
	}()

	stmt, err := tx.Prepare(`UPDATE proxies SET status=?, latency_ms=CASE WHEN ?>0 THEN ? ELSE NULL END, exit_ip=CASE WHEN ?!='' THEN ? ELSE NULL END, region=CASE WHEN ?!='' THEN ? ELSE NULL END, isp=CASE WHEN ?!='' THEN ? ELSE NULL END, last_checked_at=? WHERE id=?`)
	if err != nil { return }
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, u := range updates {
		if _, err = stmt.Exec(string(u.Status), u.Latency, u.Latency, u.ExitIP, u.ExitIP, u.Region, u.Region, u.ISP, u.ISP, now, u.ID); err != nil {
			return
		}
	}
	err = tx.Commit()
}

// ---------- Emails ----------

func (s *sqliteStore) ListEmails() []types.Email {
	rows, err := s.db.Query(`SELECT id,address,provider,password,recovery_email,paypal_id,note,status,created_at FROM emails ORDER BY created_at DESC`)
	if err != nil { return nil }
	defer rows.Close()
	var out []types.Email
	for rows.Next() {
		var e types.Email
		var provider, password, recovery, paypalID, note, createdAt sql.NullString
		_ = rows.Scan(&e.ID, &e.Address, &provider, &password, &recovery, &paypalID, &note, &e.Status, &createdAt)
		if provider.Valid { e.Provider = provider.String }
		if password.Valid { e.Password = password.String }
		if recovery.Valid { e.RecoveryEmail = recovery.String }
		if paypalID.Valid && paypalID.String != "" { e.PayPalID = &paypalID.String }
		if note.Valid { e.Note = note.String }
		if createdAt.Valid && createdAt.String != "" {
			if t, err := time.Parse(time.RFC3339Nano, createdAt.String); err == nil {
				e.CreatedAt = t
			}
		}
		out = append(out, e)
	}
	return out
}

func (s *sqliteStore) CreateEmail(e types.Email) types.Email {
	if e.ID == "" { e.ID = uuid.New().String() }
	if e.Status == "" { e.Status = "active" }
	e.CreatedAt = time.Now()
	_, _ = s.db.Exec(
		`INSERT OR IGNORE INTO emails(id,address,provider,password,recovery_email,paypal_id,note,status,created_at) VALUES(?,?,?,?,?,?,?,?,?)`,
		e.ID, e.Address, e.Provider, e.Password, e.RecoveryEmail, nullStr(e.PayPalID), e.Note, e.Status, e.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	return e
}

func (s *sqliteStore) UpdateEmail(e types.Email) (types.Email, bool) {
	res, err := s.db.Exec(
		`UPDATE emails SET address=?,provider=?,password=?,recovery_email=?,paypal_id=?,note=?,status=? WHERE id=?`,
		e.Address, e.Provider, e.Password, e.RecoveryEmail, nullStr(e.PayPalID), e.Note, e.Status, e.ID,
	)
	if err != nil { return types.Email{}, false }
	n, _ := res.RowsAffected()
	return e, n > 0
}

func (s *sqliteStore) DeleteEmail(id string) bool {
	res, _ := s.db.Exec(`DELETE FROM emails WHERE id=?`, id)
	n, _ := res.RowsAffected()
	return n > 0
}

// ---------- PayPals ----------

func (s *sqliteStore) ListPayPals() []types.PayPal {
	rows, err := s.db.Query(`SELECT id,email,owner_name,verified,balance,currency,status,note,created_at FROM paypals ORDER BY created_at DESC`)
	if err != nil { return nil }
	defer rows.Close()
	var out []types.PayPal
	for rows.Next() {
		var p types.PayPal
		var ownerName, note, createdAt sql.NullString
		_ = rows.Scan(&p.ID, &p.Email, &ownerName, &p.Verified, &p.Balance, &p.Currency, &p.Status, &note, &createdAt)
		if ownerName.Valid { p.OwnerName = ownerName.String }
		if note.Valid { p.Note = note.String }
		if createdAt.Valid && createdAt.String != "" {
			if t, err := time.Parse(time.RFC3339Nano, createdAt.String); err == nil {
				p.CreatedAt = t
			}
		}
		out = append(out, p)
	}
	return out
}

func (s *sqliteStore) CreatePayPal(p types.PayPal) types.PayPal {
	if p.ID == "" { p.ID = uuid.New().String() }
	if p.Status == "" { p.Status = "active" }
	if p.Currency == "" { p.Currency = "USD" }
	p.CreatedAt = time.Now()
	_, _ = s.db.Exec(
		`INSERT OR IGNORE INTO paypals(id,email,owner_name,verified,balance,currency,status,note,created_at) VALUES(?,?,?,?,?,?,?,?,?)`,
		p.ID, p.Email, p.OwnerName, p.Verified, p.Balance, p.Currency, p.Status, p.Note, p.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	return p
}

func (s *sqliteStore) UpdatePayPal(p types.PayPal) (types.PayPal, bool) {
	res, err := s.db.Exec(
		`UPDATE paypals SET email=?,owner_name=?,verified=?,balance=?,currency=?,status=?,note=? WHERE id=?`,
		p.Email, p.OwnerName, p.Verified, p.Balance, p.Currency, p.Status, p.Note, p.ID,
	)
	if err != nil { return types.PayPal{}, false }
	n, _ := res.RowsAffected()
	return p, n > 0
}

func (s *sqliteStore) DeletePayPal(id string) bool {
	tx, err := s.db.Begin()
	if err != nil { return false }
	// cascade: unlink emails, delete income
	_, _ = tx.Exec(`UPDATE emails SET paypal_id=NULL WHERE paypal_id=?`, id)
	_, _ = tx.Exec(`DELETE FROM income WHERE paypal_id=?`, id)
	res, err := tx.Exec(`DELETE FROM paypals WHERE id=?`, id)
	if err != nil { _ = tx.Rollback(); return false }
	n, _ := res.RowsAffected()
	_ = tx.Commit()
	return n > 0
}

// ---------- Income ----------

func (s *sqliteStore) ListIncome() []types.Income {
	rows, err := s.db.Query(`SELECT id,amount,currency,source,paypal_id,description,received_at,created_at FROM income ORDER BY received_at DESC`)
	if err != nil { return nil }
	defer rows.Close()
	var out []types.Income
	for rows.Next() {
		var i types.Income
		var source, paypalID, description sql.NullString
		var receivedAt, createdAt string
		_ = rows.Scan(&i.ID, &i.Amount, &i.Currency, &source, &paypalID, &description, &receivedAt, &createdAt)
		if source.Valid { i.Source = source.String }
		if paypalID.Valid && paypalID.String != "" { i.PayPalID = &paypalID.String }
		if description.Valid { i.Description = description.String }
		if t := parseTime(receivedAt); t != nil { i.ReceivedAt = *t }
		if t := parseTime(createdAt); t != nil { i.CreatedAt = *t }
		out = append(out, i)
	}
	return out
}

func (s *sqliteStore) CreateIncome(i types.Income) types.Income {
	if i.ID == "" { i.ID = uuid.New().String() }
	if i.Currency == "" { i.Currency = "USD" }
	if i.ReceivedAt.IsZero() { i.ReceivedAt = time.Now() }
	i.CreatedAt = time.Now()
	_, _ = s.db.Exec(
		`INSERT INTO income(id,amount,currency,source,paypal_id,description,received_at,created_at) VALUES(?,?,?,?,?,?,?,?)`,
		i.ID, i.Amount, i.Currency, i.Source, nullStr(i.PayPalID), i.Description,
		i.ReceivedAt.UTC().Format(time.RFC3339Nano), i.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	return i
}

func (s *sqliteStore) DeleteIncome(id string) bool {
	res, _ := s.db.Exec(`DELETE FROM income WHERE id=?`, id)
	n, _ := res.RowsAffected()
	return n > 0
}

// GetIncomeReport uses SQL aggregation — efficient even for millions of rows.
func (s *sqliteStore) GetIncomeReport() types.IncomeReport {
	report := types.IncomeReport{
		BySource: make(map[string]float64),
		ByMonth:  make(map[string]float64),
	}

	// Total + count
	row := s.db.QueryRow(`SELECT COALESCE(SUM(amount),0), COUNT(*) FROM income`)
	_ = row.Scan(&report.TotalUSD, &report.Count)

	// By source
	rows, err := s.db.Query(`SELECT COALESCE(source,'other') as src, SUM(amount) FROM income GROUP BY src`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var src string
			var sum float64
			_ = rows.Scan(&src, &sum)
			report.BySource[src] = sum
		}
	}

	// By month (YYYY-MM)
	rows2, err := s.db.Query(`SELECT strftime('%Y-%m', received_at) as month, SUM(amount) FROM income GROUP BY month ORDER BY month`)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var month string
			var sum float64
			_ = rows2.Scan(&month, &sum)
			report.ByMonth[month] = sum
		}
	}

	return report
}

// ---------- Nodes ----------

func (s *sqliteStore) ListNodes() []types.Node {
	rows, err := s.db.Query(`SELECT id,name,enabled,public_key,ssh_host,ssh_port,ssh_user,ssh_password,ssh_key,status,version,last_seen_at,deploy_status,deploy_log,deployed_at,email_id,lan_subnet,created_at FROM nodes ORDER BY created_at DESC`)
	if err != nil { return nil }
	defer rows.Close()
	var out []types.Node
	for rows.Next() {
		out = append(out, scanNode(rows))
	}
	return out
}

func (s *sqliteStore) CreateNode(n types.Node) types.Node {
	if n.ID == "" { n.ID = uuid.New().String() }
	if n.SSHPort == 0 { n.SSHPort = 22 }
	if n.SSHUser == "" { n.SSHUser = "root" }
	if n.Status == "" { n.Status = "offline" }
	if n.DeployStatus == "" { n.DeployStatus = "pending" }
	n.Enabled = true // default enabled
	n.CreatedAt = time.Now()
	_, _ = s.db.Exec(
		`INSERT INTO nodes(id,name,enabled,public_key,ssh_host,ssh_port,ssh_user,ssh_password,ssh_key,status,version,deploy_status,lan_subnet,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		n.ID, n.Name, n.Enabled, n.PublicKey, n.SSHHost, n.SSHPort, n.SSHUser, n.SSHPassword, n.SSHKey,
		n.Status, n.Version, n.DeployStatus, n.LanSubnet, n.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	return n
}

func (s *sqliteStore) GetNode(id string) (types.Node, bool) {
	rows, err := s.db.Query(`SELECT id,name,enabled,public_key,ssh_host,ssh_port,ssh_user,ssh_password,ssh_key,status,version,last_seen_at,deploy_status,deploy_log,deployed_at,email_id,lan_subnet,created_at FROM nodes WHERE id=?`, id)
	if err != nil { return types.Node{}, false }
	defer rows.Close()
	if !rows.Next() { return types.Node{}, false }
	return scanNode(rows), true
}

func (s *sqliteStore) UpdateNode(n types.Node) (types.Node, bool) {
	res, err := s.db.Exec(
		`UPDATE nodes SET name=?,public_key=?,ssh_host=?,ssh_port=?,ssh_user=?,ssh_password=?,ssh_key=?,email_id=?,lan_subnet=? WHERE id=?`,
		n.Name, n.PublicKey, n.SSHHost, n.SSHPort, n.SSHUser, n.SSHPassword, n.SSHKey, n.EmailID, n.LanSubnet, n.ID,
	)
	if err != nil { return types.Node{}, false }
	nn, _ := res.RowsAffected()
	if nn == 0 { return types.Node{}, false }
	updated, _ := s.GetNode(n.ID)
	return updated, true
}

func (s *sqliteStore) UpdateNodeEnabled(id string, enabled bool) bool {
	res, err := s.db.Exec(`UPDATE nodes SET enabled=? WHERE id=?`, enabled, id)
	if err != nil { return false }
	n, _ := res.RowsAffected()
	return n > 0
}

func (s *sqliteStore) DeleteNode(id string) bool {
	res, _ := s.db.Exec(`DELETE FROM nodes WHERE id=?`, id)
	n, _ := res.RowsAffected()
	return n > 0
}

func (s *sqliteStore) UpdateNodeStatus(id, status, version string, lastSeen time.Time) bool {
	res, err := s.db.Exec(
		`UPDATE nodes SET status=?, version=?, last_seen_at=? WHERE id=?`,
		status, version, lastSeen.UTC().Format(time.RFC3339Nano), id,
	)
	if err != nil { return false }
	n, _ := res.RowsAffected()
	return n > 0
}

func (s *sqliteStore) UpdateNodeDeploy(id, deployStatus, deployLog string) bool {
	var err error
	switch deployStatus {
	case "deployed":
		// Only update status + timestamp; preserve accumulated log.
		now := time.Now().UTC().Format(time.RFC3339Nano)
		_, err = s.db.Exec(`UPDATE nodes SET deploy_status=?, deployed_at=? WHERE id=?`, deployStatus, now, id)
	case "failed":
		if deployLog != "" {
			// Append error to existing log instead of overwriting.
			_, err = s.db.Exec(`UPDATE nodes SET deploy_status=?, deploy_log=deploy_log||? WHERE id=?`, deployStatus, "\n"+deployLog, id)
		} else {
			_, err = s.db.Exec(`UPDATE nodes SET deploy_status=? WHERE id=?`, deployStatus, id)
		}
	default:
		_, err = s.db.Exec(`UPDATE nodes SET deploy_status=?, deploy_log=? WHERE id=?`, deployStatus, deployLog, id)
	}
	return err == nil
}

func (s *sqliteStore) GetNodeAssignments(nodeID string) types.NodeAssignment {
	result := types.NodeAssignment{NodeID: nodeID}
	// Include both:
	// 1. Proxies assigned via mappings (client routing)
	// 2. Proxies directly assigned to this node via proxy.node_id (health-check assignments)
	rows, err := s.db.Query(`
		SELECT m.id, COALESCE(c.ip_cidr, ''), COALESCE(m.local_redirect_port, 0),
		       p.id, p.label, p.type, p.host, p.port, p.username, p.password, p.enabled, p.status, p.latency_ms, p.exit_ip, p.last_checked_at
		FROM mappings m
		JOIN clients c ON c.id = m.client_id
		JOIN proxies p ON p.id = m.proxy_id
		WHERE m.node_id = ?
		UNION
		SELECT p.id, '', 0,
		       p.id, p.label, p.type, p.host, p.port, p.username, p.password, p.enabled, p.status, p.latency_ms, p.exit_ip, p.last_checked_at
		FROM proxies p
		WHERE p.node_id = ?
		  AND p.id NOT IN (SELECT proxy_id FROM mappings WHERE node_id = ?)
		ORDER BY 1`, nodeID, nodeID, nodeID)
	if err != nil { return result }
	defer rows.Close()
	for rows.Next() {
		var pair types.ProxyMappingPair
		var pLabel, pUsername, pPassword, pExitIP, pLastChecked sql.NullString
		var pLatencyMs sql.NullInt64
		_ = rows.Scan(
			&pair.MappingID, &pair.ClientCIDR, &pair.LocalPort,
			&pair.Proxy.ID, &pLabel, &pair.Proxy.Type, &pair.Proxy.Host, &pair.Proxy.Port,
			&pUsername, &pPassword, &pair.Proxy.Enabled, &pair.Proxy.Status,
			&pLatencyMs, &pExitIP, &pLastChecked,
		)
		if pLabel.Valid { pair.Proxy.Label = pLabel.String }
		if pUsername.Valid && pUsername.String != "" { pair.Proxy.Username = &pUsername.String }
		if pPassword.Valid && pPassword.String != "" { pair.Proxy.Password = &pPassword.String }
		if pLatencyMs.Valid { v := int(pLatencyMs.Int64); pair.Proxy.LatencyMs = &v }
		if pExitIP.Valid && pExitIP.String != "" { pair.Proxy.ExitIP = &pExitIP.String }
		result.Assignments = append(result.Assignments, pair)
	}
	return result
}

// scanNode reads one row from a *sql.Rows into types.Node.
func scanNode(rows *sql.Rows) types.Node {
	var n types.Node
	var pubKey, sshPass, sshKey, version, lastSeenAt, deployLog, deployedAt, emailID, lanSubnet sql.NullString
	var createdAt string
	_ = rows.Scan(
		&n.ID, &n.Name, &n.Enabled, &pubKey, &n.SSHHost, &n.SSHPort, &n.SSHUser,
		&sshPass, &sshKey, &n.Status, &version,
		&lastSeenAt, &n.DeployStatus, &deployLog, &deployedAt, &emailID, &lanSubnet, &createdAt,
	)
	if pubKey.Valid { n.PublicKey = pubKey.String }
	if sshPass.Valid { n.SSHPassword = sshPass.String }
	if sshKey.Valid { n.SSHKey = sshKey.String }
	if version.Valid { n.Version = version.String }
	if deployLog.Valid { n.DeployLog = deployLog.String }
	if lastSeenAt.Valid { n.LastSeenAt = parseTime(lastSeenAt.String) }
	if deployedAt.Valid { n.DeployedAt = parseTime(deployedAt.String) }
	if emailID.Valid { n.EmailID = &emailID.String }
	if lanSubnet.Valid { n.LanSubnet = lanSubnet.String }
	if t := parseTime(createdAt); t != nil { n.CreatedAt = *t }
	return n
}

// Ensure sqliteStore satisfies Store interface at compile time.
var _ Store = (*sqliteStore)(nil)

// sortProxies is a helper used by ListProxies (already sorted by SQL ORDER BY).
var _ = sort.Slice // suppress unused import if needed
