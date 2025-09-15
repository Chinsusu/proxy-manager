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
		
		// Servers
		protected.GET("/servers", serverHandler.GetServers)
		protected.POST("/servers", serverHandler.CreateServer)
		protected.GET("/servers/:id", serverHandler.GetServer)
		protected.PATCH("/servers/:id", serverHandler.UpdateServer)
		protected.DELETE("/servers/:id", serverHandler.DeleteServer)
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
