package handlers

import (
	"database/sql"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"pulseops-backend/models"
)

type AuthHandler struct {
	DB        *sql.DB
	JWTSecret []byte
}

func NewAuthHandler(db *sql.DB, secret string) *AuthHandler {
	return &AuthHandler{
		DB:        db,
		JWTSecret: []byte(secret),
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Provide a valid name, email, and password of at least 8 characters",
		})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Name == "" || len(req.Name) > 100 ||
		len(req.Email) > 254 ||
		!regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`).MatchString(req.Email) ||
		len(req.Password) > 72 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid registration details",
		})
		return
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not secure password",
		})
		return
	}

	result, err := h.DB.Exec(
		`INSERT INTO users (name, email, password_hash)
		 VALUES (?, ?, ?)`,
		req.Name, req.Email, string(hash),
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Email already registered",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not create account",
		})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not retrieve account",
		})
		return
	}

	claims := jwt.MapClaims{
		"sub": strconv.Itoa(int(id)),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, _ := token.SignedString(h.JWTSecret)

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Account created successfully",
		"token":        signedToken,
		"access_token": signedToken,
		"user": models.User{
			ID:    int(id),
			Name:  req.Name,
			Email: req.Email,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Provide a valid email and password",
		})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	var user models.User
	var passwordHash string

	err := h.DB.QueryRow(
		`SELECT id, name, email, password_hash, created_at
		 FROM users WHERE email = ?`,
		req.Email,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&passwordHash,
		&user.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(req.Password),
	); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	claims := jwt.MapClaims{
		"sub": strconv.Itoa(user.ID),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(h.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not create access token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":        signedToken,
		"access_token": signedToken,
		"token_type":   "Bearer",
		"expires_in":   86400,
		"user":         user,
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var user models.User
	err := h.DB.QueryRow(
		`SELECT id, name, email, created_at FROM users WHERE id = ?`,
		userIDVal,
	).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"user_id": user.ID,
		"message": "Authentication successful",
	})
}
