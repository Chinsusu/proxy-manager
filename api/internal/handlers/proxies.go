package handlers

import (
	"net/http"
	"strconv"

	"github.com/Chinsusu/proxy-manager/api/internal/database"
	"github.com/Chinsusu/proxy-manager/api/internal/models"
	"github.com/gin-gonic/gin"
)

type ProxyHandler struct {
	db *database.DB
}

func NewProxyHandler(db *database.DB) *ProxyHandler {
	return &ProxyHandler{db: db}
}

type CreateProxyRequest struct {
	ServerID *uint  `json:"server_id,omitempty"`
	Label    string `json:"label" binding:"required"`
	Type     string `json:"type" binding:"required"`
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required,min=1,max=65535"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateProxyRequest struct {
	Label    *string `json:"label"`
	Type     *string `json:"type"`
	Host     *string `json:"host"`
	Port     *int    `json:"port"`
	Username *string `json:"username"`
	Password *string `json:"password"`
	Health   *string `json:"health"`
}

// GetServerProxies returns all proxies for a specific server
func (h *ProxyHandler) GetServerProxies(c *gin.Context) {
	serverID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	// Verify server exists
	var server models.Server
	if err := h.db.First(&server, serverID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	var proxies []models.Proxy
	result := h.db.Where("server_id = ?", serverID).Find(&proxies)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch proxies"})
		return
	}

	c.JSON(http.StatusOK, proxies)
}

// GetProxies returns all proxies (optional: can filter by server)
func (h *ProxyHandler) GetProxies(c *gin.Context) {
	var proxies []models.Proxy
	query := h.db.Preload("Server")
	
	// Optional server filter
	if serverID := c.Query("server_id"); serverID != "" {
		if id, err := strconv.ParseUint(serverID, 10, 32); err == nil {
			query = query.Where("server_id = ?", id)
		}
	}

	result := query.Find(&proxies)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch proxies"})
		return
	}

	c.JSON(http.StatusOK, proxies)
}

// GetProxy returns a single proxy by ID
func (h *ProxyHandler) GetProxy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	var proxy models.Proxy
	result := h.db.Preload("Server").First(&proxy, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
		return
	}

	c.JSON(http.StatusOK, proxy)
}

// CreateServerProxy creates a new proxy for a specific server
func (h *ProxyHandler) CreateServerProxy(c *gin.Context) {
	serverID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	var req CreateProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Override server ID from URL param
	serverIdUint := uint(serverID); req.ServerID = &serverIdUint

	// Verify server exists
	var server models.Server
	if err := h.db.First(&server, serverID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	// Validate proxy type
	validTypes := map[string]bool{"http": true, "https": true, "socks4": true, "socks5": true}
	if !validTypes[req.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy type. Must be: http, https, socks4, socks5"})
		return
	}

	proxy := models.Proxy{
		ServerID: req.ServerID,
		Label:    req.Label,
		Type:     req.Type,
		Host:     req.Host,
		Port:     req.Port,
		Username: req.Username,
		Password: req.Password,
		Health:   "unknown",
	}

	if err := h.db.Create(&proxy).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy"})
		return
	}

	// Increment config version for the server
	if req.ServerID != nil && *req.ServerID > 0 { h.db.IncrementConfigVersion(*req.ServerID) }

	// Reload proxy with server info
	h.db.Preload("Server").First(&proxy, proxy.ID)
	c.JSON(http.StatusCreated, proxy)
}

// CreateProxy creates a new proxy (requires server_id in body)
func (h *ProxyHandler) CreateProxy(c *gin.Context) {
	var req CreateProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}


	// Verify server exists (if server_id > 0)
	if req.ServerID != nil && *req.ServerID > 0 {
		var server models.Server
		if err := h.db.First(&server, *req.ServerID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
			return
		}
	}

	// Validate proxy type
	validTypes := map[string]bool{"http": true, "https": true, "socks4": true, "socks5": true}
	if !validTypes[req.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy type. Must be: http, https, socks4, socks5"})
		return
	}

	proxy := models.Proxy{
		ServerID: req.ServerID,
		Label:    req.Label,
		Type:     req.Type,
		Host:     req.Host,
		Port:     req.Port,
		Username: req.Username,
		Password: req.Password,
		Health:   "unknown",
	}

	if err := h.db.Create(&proxy).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy"})
		return
	}

	// Increment config version for the server
	if req.ServerID != nil && *req.ServerID > 0 { h.db.IncrementConfigVersion(*req.ServerID) }

	// Reload proxy with server info
	h.db.Preload("Server").First(&proxy, proxy.ID)
	c.JSON(http.StatusCreated, proxy)
}

// UpdateProxy updates an existing proxy
func (h *ProxyHandler) UpdateProxy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	var req UpdateProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	var proxy models.Proxy
	if err := h.db.First(&proxy, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
		return
	}

	// Update fields if provided
	updates := make(map[string]interface{})
	if req.Label != nil {
		updates["label"] = *req.Label
	}
	if req.Type != nil {
		validTypes := map[string]bool{"http": true, "https": true, "socks4": true, "socks5": true}
		if !validTypes[*req.Type] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy type"})
			return
		}
		updates["type"] = *req.Type
	}
	if req.Host != nil {
		updates["host"] = *req.Host
	}
	if req.Port != nil {
		if *req.Port < 1 || *req.Port > 65535 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Port must be between 1 and 65535"})
			return
		}
		updates["port"] = *req.Port
	}
	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.Password != nil {
		updates["password"] = *req.Password
	}
	if req.Health != nil {
		validHealth := map[string]bool{"ok": true, "fail": true, "unknown": true}
		if !validHealth[*req.Health] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid health status. Must be: ok, fail, unknown"})
			return
		}
		updates["health"] = *req.Health
	}

	if err := h.db.Model(&proxy).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update proxy"})
		return
	}

	// Increment config version for the server
	if proxy.ServerID != nil && *proxy.ServerID > 0 { h.db.IncrementConfigVersion(*proxy.ServerID) }

	// Reload proxy with updated data
	h.db.Preload("Server").First(&proxy, id)
	c.JSON(http.StatusOK, proxy)
}

// DeleteProxy deletes a proxy
func (h *ProxyHandler) DeleteProxy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	var proxy models.Proxy
	if err := h.db.First(&proxy, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
		return
	}

	serverID := proxy.ServerID

	if err := h.db.Delete(&proxy).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete proxy"})
		return
	}

	// Increment config version for the server
	if serverID != nil && *serverID > 0 { h.db.IncrementConfigVersion(*serverID) }

	c.JSON(http.StatusOK, gin.H{"message": "Proxy deleted successfully"})
}

// MoveProxyRequest for moving single proxy to group
type MoveProxyRequest struct {
	GroupID *uint `json:"group_id"`
}

// BulkMoveProxyRequest for moving multiple proxies to group
type BulkMoveProxyRequest struct {
	ProxyIDs []uint `json:"proxy_ids" binding:"required"`
	GroupID  *uint  `json:"group_id"`
}

// MoveProxyToGroup moves a proxy to a specific group
func (h *ProxyHandler) MoveProxyToGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	var req MoveProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Check if proxy exists
	var proxy models.Proxy
	if err := h.db.First(&proxy, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
		return
	}

	// Check if group exists (if group_id is provided)
	if req.GroupID != nil && *req.GroupID > 0 {
		var group models.ProxyGroup
		if err := h.db.First(&group, *req.GroupID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
			return
		}
	}

	// Update proxy's group_id
	if err := h.db.Model(&proxy).Update("group_id", req.GroupID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move proxy"})
		return
	}

	// Increment config version for the server
	if proxy.ServerID != nil && *proxy.ServerID > 0 {
		h.db.IncrementConfigVersion(*proxy.ServerID)
	}

	// Reload proxy with updated data
	h.db.Preload("Server").Preload("Group").First(&proxy, id)
	c.JSON(http.StatusOK, proxy)
}

// BulkMoveProxiesToGroup moves multiple proxies to a specific group
func (h *ProxyHandler) BulkMoveProxiesToGroup(c *gin.Context) {
	var req BulkMoveProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	if len(req.ProxyIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No proxy IDs provided"})
		return
	}

	// Check if group exists (if group_id is provided)
	if req.GroupID != nil && *req.GroupID > 0 {
		var group models.ProxyGroup
		if err := h.db.First(&group, *req.GroupID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
			return
		}
	}

	// Get all proxies to be moved
	var proxies []models.Proxy
	if err := h.db.Where("id IN ?", req.ProxyIDs).Find(&proxies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find proxies"})
		return
	}

	if len(proxies) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No proxies found with provided IDs"})
		return
	}

	// Update all proxies' group_id in bulk
	if err := h.db.Model(&models.Proxy{}).Where("id IN ?", req.ProxyIDs).Update("group_id", req.GroupID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move proxies"})
		return
	}

	// Increment config version for all affected servers
	serverIDs := make(map[uint]bool)
	for _, proxy := range proxies {
		if proxy.ServerID != nil && *proxy.ServerID > 0 {
			serverIDs[*proxy.ServerID] = true
		}
	}
	
	for serverID := range serverIDs {
		h.db.IncrementConfigVersion(serverID)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Proxies moved successfully",
		"moved_count": len(proxies),
	})
}
