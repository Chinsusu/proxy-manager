package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	
	"github.com/Chinsusu/proxy-manager/api/internal/database"
	"github.com/Chinsusu/proxy-manager/api/internal/models"
)

type GroupHandler struct {
	db *database.DB
}

func NewGroupHandler(db *database.DB) *GroupHandler {
	return &GroupHandler{db: db}
}

// GET /groups
func (h *GroupHandler) GetGroups(c *gin.Context) {
	var groups []models.ProxyGroup
	
	if err := h.db.Preload("Proxies").Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}
	
	c.JSON(http.StatusOK, groups)
}

// POST /groups
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var group models.ProxyGroup
	
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.db.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}
	
	c.JSON(http.StatusCreated, group)
}

// PUT /groups/:id
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}
	
	var group models.ProxyGroup
	if err := h.db.First(&group, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}
	
	var updateData models.ProxyGroup
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	group.Name = updateData.Name
	group.Description = updateData.Description
	
	if err := h.db.Save(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group"})
		return
	}
	
	c.JSON(http.StatusOK, group)
}

// DELETE /groups/:id
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}
	
	// Check if group exists
	var group models.ProxyGroup
	if err := h.db.First(&group, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}
	
	// Check if group has proxies
	var proxyCount int64
	h.db.Model(&models.Proxy{}).Where("group_id = ?", id).Count(&proxyCount)
	if proxyCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete group with proxies. Move proxies to another group first."})
		return
	}
	
	if err := h.db.Delete(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete group"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}
