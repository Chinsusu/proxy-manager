package handlers

import (
	"net/http"
	"time"

	"github.com/Chinsusu/proxy-manager/api/internal/database"
	"github.com/Chinsusu/proxy-manager/api/internal/models"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	db *database.DB
}

func NewAdminHandler(db *database.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

// Health returns API health status
func (h *AdminHandler) Health(c *gin.Context) {
	// Test database connection
	sqlDB, err := h.db.DB.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "error",
			"timestamp": time.Now(),
			"error":     "database connection failed",
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "error",
			"timestamp": time.Now(),
			"error":     "database ping failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now(),
	})
}

// Summary returns system summary
func (h *AdminHandler) Summary(c *gin.Context) {
	var (
		serverCount       int64
		proxyCount        int64
		mappingCount      int64
		activeServerCount int64
	)

	// Count total servers
	h.db.Model(&models.Server{}).Count(&serverCount)
	
	// Count total proxies
	h.db.Model(&models.Proxy{}).Count(&proxyCount)
	
	// Count total mappings
	h.db.Model(&models.Mapping{}).Count(&mappingCount)
	
	// Count active servers (last seen within 5 minutes)
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	h.db.Model(&models.Server{}).
		Where("last_seen_at > ? OR status = ?", fiveMinutesAgo, "online").
		Count(&activeServerCount)

	c.JSON(http.StatusOK, gin.H{
		"servers":        serverCount,
		"proxies":        proxyCount,
		"mappings":       mappingCount,
		"active_servers": activeServerCount,
		"timestamp":      time.Now(),
	})
}
