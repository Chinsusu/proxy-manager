package models

import (
	"time"
	"gorm.io/gorm"
)

// User represents admin user
type User struct {
	ID           uint      `json:"id" gorm:"primarykey"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Role         string    `json:"role" gorm:"default:admin"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ProxyGroup represents a group of proxies
type ProxyGroup struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	Name        string    `json:"name" gorm:"not null;uniqueIndex"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// Relationships
	Proxies []Proxy `json:"proxies,omitempty" gorm:"foreignKey:GroupID"`
}

// Server represents a proxy server node/agent
type Server struct {
	ID           uint      `json:"id" gorm:"primarykey"`
	Name         string    `json:"name" gorm:"not null"`
	Tags         string    `json:"tags"` // JSON array as string
	WANIface     string    `json:"wan_iface"`
	LANIface     string    `json:"lan_iface"`
	LastSeenAt   *time.Time `json:"last_seen_at"`
	Status       string    `json:"status" gorm:"default:offline"` // online/offline
	AgentToken   string    `json:"-" gorm:"uniqueIndex;not null"` // for agent auth
	ConfigVersion int      `json:"config_version" gorm:"default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	
	// Relationships
	Proxies  []Proxy   `json:"proxies,omitempty"`
	Mappings []Mapping `json:"mappings,omitempty"`
}

// Proxy represents upstream proxy server
type Proxy struct {
	ID         uint      `json:"id" gorm:"primarykey"`
	ServerID   *uint     `json:"server_id"`
	GroupID    *uint     `json:"group_id"`
	Label      string    `json:"label" gorm:"not null"`
	Type       string    `json:"type" gorm:"not null"` // http, socks5, etc.
	Host       string    `json:"host" gorm:"not null"`
	Port       int       `json:"port" gorm:"not null"`
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	Health     string    `json:"health" gorm:"default:unknown"` // ok, fail, unknown
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	
	// Relationships
	Server   Server     `json:"server,omitempty"`
	Group    ProxyGroup `json:"group,omitempty"`
	Mappings []Mapping  `json:"mappings,omitempty" gorm:"foreignKey:UpstreamProxyID"`
}

// Mapping represents client to proxy mapping rules
type Mapping struct {
	ID               uint      `json:"id" gorm:"primarykey"`
	ServerID         uint      `json:"server_id" gorm:"not null"`
	ClientCIDR       string    `json:"client_cidr" gorm:"not null"`
	DstPorts         string    `json:"dst_ports"` // JSON array as string
	UpstreamProxyID  uint      `json:"upstream_proxy_id" gorm:"not null"`
	Enabled          bool      `json:"enabled" gorm:"default:true"`
	Notes            string    `json:"notes"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	
	// Relationships
	Server        Server `json:"server,omitempty"`
	UpstreamProxy Proxy  `json:"upstream_proxy,omitempty"`
}

// AuditLog represents system audit trail
type AuditLog struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	Actor     string    `json:"actor" gorm:"not null"` // user email or "system"
	Action    string    `json:"action" gorm:"not null"` // create, update, delete
	Resource  string    `json:"resource" gorm:"not null"` // server, proxy, mapping
	Before    string    `json:"before"` // JSON
	After     string    `json:"after"` // JSON
	CreatedAt time.Time `json:"created_at"`
}

// AgentPullResponse represents response for agent pull
type AgentPullResponse struct {
	Version  int       `json:"version"`
	Proxies  []Proxy   `json:"proxies"`
	Mappings []Mapping `json:"mappings"`
}

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&ProxyGroup{},
		&Server{},
		&Proxy{},
		&Mapping{},
		&AuditLog{},
	)
}
