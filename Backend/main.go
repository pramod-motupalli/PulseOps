package main

import (
	"bufio"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"pulseops-backend/database"
	"pulseops-backend/handlers"
	authmiddleware "pulseops-backend/middleware"
)

func loadEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func main() {
	loadEnv(".env")

	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		log.Fatal("JWT_SECRET must contain at least 32 characters")
	}

	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Database initialization failed: ", err)
	}
	defer db.Close()

	authHandler := handlers.NewAuthHandler(db, secret)
	monitorHandler := handlers.NewMonitorHandler(db)

	go monitorHandler.StartBackgroundChecker(45 * time.Second)

	router := gin.Default()
	router.Use(corsMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "PulseOps",
			"status":  "healthy",
		})
	})

	router.POST("/login", authHandler.Login)
	router.POST("/register", authHandler.Register)

	api := router.Group("/api/v1")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
	}

	protected := router.Group("/")
	protected.Use(authmiddleware.AuthRequired([]byte(secret)))
	{
		protected.GET("/me", authHandler.Me)
		protected.GET("/nodes", monitorHandler.GetNodes)
		protected.POST("/nodes", monitorHandler.AddNode)
		protected.POST("/nodes/test", monitorHandler.TestEndpoint)
		protected.POST("/nodes/check-all", monitorHandler.CheckAllNodes)
		protected.POST("/nodes/:id/check", monitorHandler.CheckNode)
		protected.DELETE("/nodes/:id", monitorHandler.DeleteNode)
		protected.GET("/dashboard/stats", monitorHandler.GetDashboardStats)
		protected.GET("/alerts", monitorHandler.GetAlerts)

		v1Protected := protected.Group("/api/v1")
		{
			v1Protected.GET("/me", authHandler.Me)
			v1Protected.GET("/nodes", monitorHandler.GetNodes)
			v1Protected.POST("/nodes", monitorHandler.AddNode)
			v1Protected.POST("/nodes/test", monitorHandler.TestEndpoint)
			v1Protected.POST("/nodes/check-all", monitorHandler.CheckAllNodes)
			v1Protected.POST("/nodes/:id/check", monitorHandler.CheckNode)
			v1Protected.DELETE("/nodes/:id", monitorHandler.DeleteNode)
			v1Protected.GET("/dashboard/stats", monitorHandler.GetDashboardStats)
			v1Protected.GET("/alerts", monitorHandler.GetAlerts)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	log.Printf("PulseOps server starting on %s\n", port)
	if err := router.Run(port); err != nil {
		log.Fatal(err)
	}
}
