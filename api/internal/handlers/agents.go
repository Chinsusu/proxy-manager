package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Chinsusu/proxy-manager/api/internal/database"
	"github.com/Chinsusu/proxy-manager/api/internal/models"
	"github.com/gin-gonic/gin"
)

type AgentHandler struct {
	db *database.DB
}

func NewAgentHandler(db *database.DB) *AgentHandler {
	return &AgentHandler{db: db}
}

// Pull handles agent configuration pull requests
func (h *AgentHandler) Pull(c *gin.Context) {
	agentID := c.Param("agent_id")
	sinceVersion := c.DefaultQuery("since", "0")
	agentToken := c.GetString("agent_token")

	// Parse agent ID and since version
	serverID, err := strconv.ParseUint(agentID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	since, err := strconv.Atoi(sinceVersion)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid since version"})
		return
	}

	// Verify agent token
	var server models.Server
	if err := h.db.Where("id = ? AND agent_token = ?", serverID, agentToken).First(&server).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid agent token"})
		return
	}

	// Update last seen timestamp
	now := time.Now()
	h.db.Model(&server).Updates(map[string]interface{}{
		"last_seen_at": &now,
		"status":       "online",
	})

	// Check if config has changed
	currentVersion := server.ConfigVersion
	if since >= currentVersion {
		// No changes
		c.Status(http.StatusNoContent)
		return
	}

	// Fetch proxies and mappings for this server
	var proxies []models.Proxy
	var mappings []models.Mapping

	h.db.Where("server_id = ?", serverID).Find(&proxies)
	h.db.Preload("UpstreamProxy").Where("server_id = ?", serverID).Find(&mappings)

	response := models.AgentPullResponse{
		Version:  currentVersion,
		Proxies:  proxies,
		Mappings: mappings,
	}

	c.JSON(http.StatusOK, response)
}

// Ack handles agent acknowledgment of config application
func (h *AgentHandler) Ack(c *gin.Context) {
	agentID := c.Param("agent_id")
	agentToken := c.GetString("agent_token")

	type AckRequest struct {
		Version int    `json:"version" binding:"required"`
		Status  string `json:"status" binding:"required"`
	}

	var req AckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Parse agent ID
	serverID, err := strconv.ParseUint(agentID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	// Verify agent token
	var server models.Server
	if err := h.db.Where("id = ? AND agent_token = ?", serverID, agentToken).First(&server).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid agent token"})
		return
	}

	// Log the acknowledgment (could be stored in audit log)
	// For now, just update last seen
	now := time.Now()
	h.db.Model(&server).Updates(map[string]interface{}{
		"last_seen_at": &now,
		"status":       "online",
	})

	c.JSON(http.StatusOK, gin.H{"message": "Acknowledgment received"})
}
