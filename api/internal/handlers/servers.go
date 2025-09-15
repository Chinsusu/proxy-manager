package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/Chinsusu/proxy-manager/api/internal/database"
	"github.com/Chinsusu/proxy-manager/api/internal/models"
	"github.com/gin-gonic/gin"
)

type ServerHandler struct {
	db *database.DB
}

func NewServerHandler(db *database.DB) *ServerHandler {
	return &ServerHandler{db: db}
}

type CreateServerRequest struct {
	Name     string   `json:"name" binding:"required"`
	Tags     []string `json:"tags"`
	WANIface string   `json:"wan_iface"`
	LANIface string   `json:"lan_iface"`
}

type UpdateServerRequest struct {
	Name     *string   `json:"name"`
	Tags     *[]string `json:"tags"`
	WANIface *string   `json:"wan_iface"`
	LANIface *string   `json:"lan_iface"`
	Status   *string   `json:"status"`
}

// GetServers returns all servers
func (h *ServerHandler) GetServers(c *gin.Context) {
	var servers []models.Server
	
	result := h.db.Preload("Proxies").Preload("Mappings").Find(&servers)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch servers"})
		return
	}

	c.JSON(http.StatusOK, servers)
}

// GetServer returns a single server by ID
func (h *ServerHandler) GetServer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	var server models.Server
	result := h.db.Preload("Proxies").Preload("Mappings").First(&server, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	c.JSON(http.StatusOK, server)
}

// CreateServer creates a new server
func (h *ServerHandler) CreateServer(c *gin.Context) {
	var req CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Generate agent token
	agentToken, err := generateAgentToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate agent token"})
		return
	}

	// Convert tags to JSON string
	tagsJSON := "[]"
	if len(req.Tags) > 0 {
		// Simple JSON encoding for tags
		tagsJSON = `["` + req.Tags[0]
		for i := 1; i < len(req.Tags); i++ {
			tagsJSON += `","` + req.Tags[i]
		}
		tagsJSON += `"]`
	}

	server := models.Server{
		Name:          req.Name,
		Tags:          tagsJSON,
		WANIface:      req.WANIface,
		LANIface:      req.LANIface,
		Status:        "offline",
		AgentToken:    agentToken,
		ConfigVersion: 0,
	}

	if err := h.db.Create(&server).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server"})
		return
	}

	c.JSON(http.StatusCreated, server)
}

// UpdateServer updates an existing server
func (h *ServerHandler) UpdateServer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	var req UpdateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	var server models.Server
	if err := h.db.First(&server, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	// Update fields if provided
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Tags != nil {
		// Convert tags to JSON string
		tagsJSON := "[]"
		if len(*req.Tags) > 0 {
			tagsJSON = `["` + (*req.Tags)[0]
			for i := 1; i < len(*req.Tags); i++ {
				tagsJSON += `","` + (*req.Tags)[i]
			}
			tagsJSON += `"]`
		}
		updates["tags"] = tagsJSON
	}
	if req.WANIface != nil {
		updates["wan_iface"] = *req.WANIface
	}
	if req.LANIface != nil {
		updates["lan_iface"] = *req.LANIface
	}
	if req.Status != nil {
		updates["status"] = *req.Status
		if *req.Status == "online" {
			now := time.Now()
			updates["last_seen_at"] = &now
		}
	}

	if err := h.db.Model(&server).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server"})
		return
	}

	// Reload server with updated data
	h.db.Preload("Proxies").Preload("Mappings").First(&server, id)
	c.JSON(http.StatusOK, server)
}

// DeleteServer deletes a server
func (h *ServerHandler) DeleteServer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	// Check if server exists
	var server models.Server
	if err := h.db.First(&server, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	// Start transaction
	tx := h.db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Delete all mappings associated with this server
	if err := tx.Unscoped().Where("server_id = ?", id).Delete(&models.Mapping{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete server mappings"})
		return
	}

	// 2. Update proxies to remove server association using raw SQL
	if err := tx.Exec("UPDATE proxies SET server_id = NULL WHERE server_id = ?", id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unassign proxies"})
		return
	}

	// 3. Delete the server itself using raw SQL to handle constraints
	if err := tx.Exec("DELETE FROM servers WHERE id = ?", id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete server"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Server deleted successfully"})
}

// generateAgentToken generates a random token for agent authentication
func generateAgentToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
