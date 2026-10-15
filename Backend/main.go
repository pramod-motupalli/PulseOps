package main

import (
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
	data, _ := os.ReadFile(filepath)
	for _, line := range strings.Split(string(data), "\n") {
		if k, v, ok := strings.Cut(strings.TrimSpace(line), "="); ok && !strings.HasPrefix(k, "#") {
			if k = strings.TrimSpace(k); k != "" && os.Getenv(k) == "" {
				os.Setenv(k, strings.Trim(strings.TrimSpace(v), `"'`))
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

	// Public routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "PulseOps", "status": "healthy"})
	})
	router.POST("/login", authHandler.Login)
	router.POST("/register", authHandler.Register)

	// Protected routes
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
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = ":5000"
	} else if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	log.Printf("PulseOps server starting on %s\n", port)
	if err := router.Run(port); err != nil {
		log.Fatal(err)
	}
}
