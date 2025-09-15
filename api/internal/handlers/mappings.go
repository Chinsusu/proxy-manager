package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Chinsusu/proxy-manager/api/internal/database"
	"github.com/Chinsusu/proxy-manager/api/internal/models"
	"github.com/gin-gonic/gin"
)

type MappingHandler struct {
	db *database.DB
}

func NewMappingHandler(db *database.DB) *MappingHandler {
	return &MappingHandler{db: db}
}

type CreateMappingRequest struct {
	ServerID        uint     `json:"server_id" binding:"required"`
	ClientCIDR      string   `json:"client_cidr" binding:"required"`
	DstPorts        []int    `json:"dst_ports" binding:"required"`
	UpstreamProxyID uint     `json:"upstream_proxy_id" binding:"required"`
	Enabled         *bool    `json:"enabled"`
	Notes           string   `json:"notes"`
}

type UpdateMappingRequest struct {
	ClientCIDR      *string  `json:"client_cidr"`
	DstPorts        *[]int   `json:"dst_ports"`
	UpstreamProxyID *uint    `json:"upstream_proxy_id"`
	Enabled         *bool    `json:"enabled"`
	Notes           *string  `json:"notes"`
}

// GetServerMappings returns all mappings for a specific server
func (h *MappingHandler) GetServerMappings(c *gin.Context) {
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

	var mappings []models.Mapping
	result := h.db.Preload("UpstreamProxy").Where("server_id = ?", serverID).Find(&mappings)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch mappings"})
		return
	}

	c.JSON(http.StatusOK, mappings)
}

// GetMappings returns all mappings (optional: can filter by server)
func (h *MappingHandler) GetMappings(c *gin.Context) {
	var mappings []models.Mapping
	query := h.db.Preload("Server").Preload("UpstreamProxy")
	
	// Optional server filter
	if serverID := c.Query("server_id"); serverID != "" {
		if id, err := strconv.ParseUint(serverID, 10, 32); err == nil {
			query = query.Where("server_id = ?", id)
		}
	}

	result := query.Find(&mappings)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch mappings"})
		return
	}

	c.JSON(http.StatusOK, mappings)
}

// GetMapping returns a single mapping by ID
func (h *MappingHandler) GetMapping(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mapping ID"})
		return
	}

	var mapping models.Mapping
	result := h.db.Preload("Server").Preload("UpstreamProxy").First(&mapping, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mapping not found"})
		return
	}

	c.JSON(http.StatusOK, mapping)
}

// CreateServerMapping creates a new mapping for a specific server
func (h *MappingHandler) CreateServerMapping(c *gin.Context) {
	serverID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	var req CreateMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Override server ID from URL param
	req.ServerID = uint(serverID)

	// Verify server exists
	var server models.Server
	if err := h.db.First(&server, serverID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	// Verify upstream proxy exists and belongs to the same server
	var proxy models.Proxy
	if err := h.db.Where("id = ? AND server_id = ?", req.UpstreamProxyID, req.ServerID).First(&proxy).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Upstream proxy not found or belongs to different server"})
		return
	}

	// Validate dst_ports
	if len(req.DstPorts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dst_ports cannot be empty"})
		return
	}
	
	for _, port := range req.DstPorts {
		if port < 1 || port > 65535 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "All ports must be between 1 and 65535"})
			return
		}
	}

	// Convert dst_ports to JSON string
	dstPortsJSON, err := json.Marshal(req.DstPorts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize dst_ports"})
		return
	}

	// Default enabled to true if not provided
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	mapping := models.Mapping{
		ServerID:        req.ServerID,
		ClientCIDR:      req.ClientCIDR,
		DstPorts:        string(dstPortsJSON),
		UpstreamProxyID: req.UpstreamProxyID,
		Enabled:         enabled,
		Notes:           req.Notes,
	}

	if err := h.db.Create(&mapping).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create mapping"})
		return
	}

	// Increment config version for the server
	h.db.IncrementConfigVersion(req.ServerID)

	// Reload mapping with relations
	h.db.Preload("Server").Preload("UpstreamProxy").First(&mapping, mapping.ID)
	c.JSON(http.StatusCreated, mapping)
}

// CreateMapping creates a new mapping (requires server_id in body)
func (h *MappingHandler) CreateMapping(c *gin.Context) {
	var req CreateMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Verify server exists
	var server models.Server
	if err := h.db.First(&server, req.ServerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	// Verify upstream proxy exists and belongs to the same server
	var proxy models.Proxy
	if err := h.db.Where("id = ? AND server_id = ?", req.UpstreamProxyID, req.ServerID).First(&proxy).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Upstream proxy not found or belongs to different server"})
		return
	}

	// Validate dst_ports
	if len(req.DstPorts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dst_ports cannot be empty"})
		return
	}
	
	for _, port := range req.DstPorts {
		if port < 1 || port > 65535 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "All ports must be between 1 and 65535"})
			return
		}
	}

	// Convert dst_ports to JSON string
	dstPortsJSON, err := json.Marshal(req.DstPorts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize dst_ports"})
		return
	}

	// Default enabled to true if not provided
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	mapping := models.Mapping{
		ServerID:        req.ServerID,
		ClientCIDR:      req.ClientCIDR,
		DstPorts:        string(dstPortsJSON),
		UpstreamProxyID: req.UpstreamProxyID,
		Enabled:         enabled,
		Notes:           req.Notes,
	}

	if err := h.db.Create(&mapping).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create mapping"})
		return
	}

	// Increment config version for the server
	h.db.IncrementConfigVersion(req.ServerID)

	// Reload mapping with relations
	h.db.Preload("Server").Preload("UpstreamProxy").First(&mapping, mapping.ID)
	c.JSON(http.StatusCreated, mapping)
}

// UpdateMapping updates an existing mapping
func (h *MappingHandler) UpdateMapping(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mapping ID"})
		return
	}

	var req UpdateMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	var mapping models.Mapping
	if err := h.db.First(&mapping, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mapping not found"})
		return
	}

	// Update fields if provided
	updates := make(map[string]interface{})
	
	if req.ClientCIDR != nil {
		updates["client_cidr"] = *req.ClientCIDR
	}
	
	if req.DstPorts != nil {
		if len(*req.DstPorts) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "dst_ports cannot be empty"})
			return
		}
		
		for _, port := range *req.DstPorts {
			if port < 1 || port > 65535 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "All ports must be between 1 and 65535"})
				return
			}
		}
		
		dstPortsJSON, err := json.Marshal(*req.DstPorts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize dst_ports"})
			return
		}
		updates["dst_ports"] = string(dstPortsJSON)
	}
	
	if req.UpstreamProxyID != nil {
		// Verify upstream proxy exists and belongs to the same server
		var proxy models.Proxy
		if err := h.db.Where("id = ? AND server_id = ?", *req.UpstreamProxyID, mapping.ServerID).First(&proxy).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Upstream proxy not found or belongs to different server"})
			return
		}
		updates["upstream_proxy_id"] = *req.UpstreamProxyID
	}
	
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	if err := h.db.Model(&mapping).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update mapping"})
		return
	}

	// Increment config version for the server
	h.db.IncrementConfigVersion(mapping.ServerID)

	// Reload mapping with updated data
	h.db.Preload("Server").Preload("UpstreamProxy").First(&mapping, id)
	c.JSON(http.StatusOK, mapping)
}

// DeleteMapping deletes a mapping
func (h *MappingHandler) DeleteMapping(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mapping ID"})
		return
	}

	var mapping models.Mapping
	if err := h.db.First(&mapping, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mapping not found"})
		return
	}

	serverID := mapping.ServerID

	if err := h.db.Delete(&mapping).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete mapping"})
		return
	}

	// Increment config version for the server
	h.db.IncrementConfigVersion(serverID)

	c.JSON(http.StatusOK, gin.H{"message": "Mapping deleted successfully"})
}
