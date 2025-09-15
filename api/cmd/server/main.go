package main

import (
	"log"

	"github.com/Chinsusu/proxy-manager/api/internal/config"
	"github.com/Chinsusu/proxy-manager/api/internal/database"
	"github.com/Chinsusu/proxy-manager/api/internal/handlers"
	"github.com/Chinsusu/proxy-manager/api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg)
	adminHandler := handlers.NewAdminHandler(db)
	serverHandler := handlers.NewServerHandler(db)
	proxyHandler := handlers.NewProxyHandler(db)
	mappingHandler := handlers.NewMappingHandler(db)
	agentHandler := handlers.NewAgentHandler(db)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Global middlewares
	r.Use(middleware.CORS())

	// API v1 routes
	v1 := r.Group("/api/v1")
	
	// Public routes (no auth required)
	auth := v1.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
	}
	
	admin := v1.Group("/admin")
	{
		admin.GET("/health", adminHandler.Health)
	}

	// Protected routes (JWT auth required)
	protected := v1.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		// Auth
		protected.GET("/auth/me", authHandler.Me)
		
		// Admin with auth
		protected.GET("/admin/summary", adminHandler.Summary)
		
		// Servers - base CRUD (avoid conflict by putting specific routes first)
		servers := protected.Group("/servers")
		{
			servers.GET("", serverHandler.GetServers)
			servers.POST("", serverHandler.CreateServer)
			
			// Server-specific sub-resources (more specific)
			servers.GET("/:server_id/proxies", proxyHandler.GetServerProxies)
			servers.POST("/:server_id/proxies", proxyHandler.CreateServerProxy)
			servers.GET("/:server_id/mappings", mappingHandler.GetServerMappings)
			servers.POST("/:server_id/mappings", mappingHandler.CreateServerMapping)
			
			// Server CRUD (less specific, put after sub-resources)
			servers.GET("/:id", serverHandler.GetServer)
			servers.PATCH("/:id", serverHandler.UpdateServer)
			servers.DELETE("/:id", serverHandler.DeleteServer)
		}
		
		// Global Proxies
		proxies := protected.Group("/proxies")
		{
			proxies.GET("", proxyHandler.GetProxies)
			proxies.POST("", proxyHandler.CreateProxy)
			proxies.GET("/:id", proxyHandler.GetProxy)
			proxies.PATCH("/:id", proxyHandler.UpdateProxy)
			proxies.DELETE("/:id", proxyHandler.DeleteProxy)
		}
		
		// Global Mappings
		mappings := protected.Group("/mappings")
		{
			mappings.GET("", mappingHandler.GetMappings)
			mappings.POST("", mappingHandler.CreateMapping)
			mappings.GET("/:id", mappingHandler.GetMapping)
			mappings.PATCH("/:id", mappingHandler.UpdateMapping)
			mappings.DELETE("/:id", mappingHandler.DeleteMapping)
		}
	}

	// Agent routes (agent token auth)
	agents := v1.Group("/agents")
	agents.Use(middleware.AgentAuth())
	{
		agents.GET("/:agent_id/pull", agentHandler.Pull)
		agents.POST("/:agent_id/ack", agentHandler.Ack)
	}

	log.Printf("Server starting on %s", cfg.APIBind)
	log.Fatal(r.Run(cfg.APIBind))
}
