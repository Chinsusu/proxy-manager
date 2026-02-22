package types

import "time"

type ProxyStatus string

const (
	StatusOK       ProxyStatus = "OK"
	StatusDegraded ProxyStatus = "DEGRADED"
	StatusDown     ProxyStatus = "DOWN"
)

// ---------- Proxy / Client / Mapping ----------

type Proxy struct {
	ID            string      `json:"id"`
	Label         string      `json:"label,omitempty"`
	Type          string      `json:"type"` // "http" | "socks5"
	Host          string      `json:"host"`
	Port          int         `json:"port"`
	Username      *string     `json:"username,omitempty"`
	Password      *string     `json:"password,omitempty"`
	Enabled       bool        `json:"enabled"`
	Status        ProxyStatus `json:"status"`
	LatencyMs     *int        `json:"latency_ms,omitempty"`
	ExitIP        *string     `json:"exit_ip,omitempty"`
	Region        *string     `json:"region,omitempty"`
	ISP           *string     `json:"isp,omitempty"`
	LastCheckedAt *time.Time  `json:"last_checked_at,omitempty"`
	NodeID        *string     `json:"node_id,omitempty"`
}

type Client struct {
	ID      string `json:"id"`
	IPCidr  string `json:"ip_cidr"`
	Note    string `json:"note,omitempty"`
	Enabled bool   `json:"enabled"`
}

type Mapping struct {
	ID                string     `json:"id"`
	ClientID          string     `json:"client_id"`
	ProxyID           string     `json:"proxy_id"`
	NodeID            *string    `json:"node_id,omitempty"`
	Protocol          string     `json:"protocol"` // "http" | "socks5"
	LocalRedirectPort int        `json:"local_redirect_port"`
	State             string     `json:"state"` // "APPLIED" | "PENDING" | "FAILED"
	LastAppliedAt     *time.Time `json:"last_applied_at,omitempty"`
}

type MappingView struct {
	ID                string  `json:"id"`
	Client            Client  `json:"client"`
	Proxy             Proxy   `json:"proxy"`
	NodeID            *string `json:"node_id,omitempty"`
	State             string  `json:"state"`
	LocalRedirectPort int     `json:"local_redirect_port"`
}

// ---------- Email Management ----------

type Email struct {
	ID            string    `json:"id"`
	Address       string    `json:"address"`
	Provider      string    `json:"provider"`                // "gmail" | "outlook" | "yahoo" | "other"
	Password      string    `json:"password,omitempty"`
	RecoveryEmail string    `json:"recovery_email,omitempty"`
	PayPalID      *string   `json:"paypal_id,omitempty"` // linked PayPal account
	Note          string    `json:"note,omitempty"`
	Status        string    `json:"status"` // "active" | "disabled" | "banned"
	CreatedAt     time.Time `json:"created_at"`
}

// ---------- PayPal Management ----------

type PayPal struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`                // PayPal login email
	OwnerName string    `json:"owner_name,omitempty"`
	Verified  bool      `json:"verified"`
	Balance   float64   `json:"balance"`  // manually entered
	Currency  string    `json:"currency"` // "USD" | "EUR" etc.
	Status    string    `json:"status"`   // "active" | "limited" | "suspended"
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ---------- Income Tracking ----------

type Income struct {
	ID          string    `json:"id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`            // "USD" | "EUR" etc.
	Source      string    `json:"source"`              // "paypal" | "crypto" | "bank" | "other"
	PayPalID    *string   `json:"paypal_id,omitempty"`
	Description string    `json:"description,omitempty"`
	ReceivedAt  time.Time `json:"received_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type IncomeReport struct {
	TotalUSD float64            `json:"total_usd"`
	BySource map[string]float64 `json:"by_source"`
	ByMonth  map[string]float64 `json:"by_month"` // "2026-02" -> sum USD
	Count    int                `json:"count"`
}

// ---------- Node Management ----------

type Node struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	PublicKey    string     `json:"public_key,omitempty"`
	SSHHost      string     `json:"ssh_host"`
	SSHPort      int        `json:"ssh_port"`
	SSHUser      string     `json:"ssh_user"`
	SSHPassword  string     `json:"ssh_password,omitempty"` // AES-256-GCM encrypted at rest
	SSHKey       string     `json:"ssh_key,omitempty"`      // AES-256-GCM encrypted at rest
	Status       string     `json:"status"`                 // "online" | "offline" | "deploying" | "error"
	Version      string     `json:"version,omitempty"`      // pgw-node version
	LastSeenAt   *time.Time `json:"last_seen_at,omitempty"`
	DeployStatus string     `json:"deploy_status"` // "pending" | "deploying" | "deployed" | "failed"
	DeployLog    string     `json:"deploy_log,omitempty"`
	DeployedAt   *time.Time `json:"deployed_at,omitempty"`
	EmailID      *string    `json:"email_id,omitempty"` // assigned email account
	CreatedAt    time.Time  `json:"created_at"`
}

// NodeAssignment is returned to node agents via GET /v1/nodes/{id}/assignments
type NodeAssignment struct {
	NodeID      string             `json:"node_id"`
	Assignments []ProxyMappingPair `json:"assignments"`
}

type ProxyMappingPair struct {
	MappingID  string `json:"mapping_id"`
	ClientCIDR string `json:"client_cidr"`
	LocalPort  int    `json:"local_port"`
	Proxy      Proxy  `json:"proxy"`
}

// NodeHeartbeat is sent from node agent via POST /v1/nodes/{id}/heartbeat
type NodeHeartbeat struct {
	Version  string                `json:"version"`
	Mappings []MappingStatusReport `json:"mappings"`
}

type MappingStatusReport struct {
	MappingID   string `json:"mapping_id"`
	ProxyID     string `json:"proxy_id"`
	State       string `json:"state"`
	ProxyStatus string `json:"proxy_status"`
	LatencyMs   int    `json:"latency_ms"`
	ExitIP      string `json:"exit_ip"`
}
