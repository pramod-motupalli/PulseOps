package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			strings.TrimSpace(parts[1]) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing or invalid authorization header",
			})
			return
		}

		claims := jwt.MapClaims{}

		token, err := jwt.ParseWithClaims(
			parts[1],
			claims,
			func(token *jwt.Token) (any, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, jwt.ErrSignatureInvalid
				}
				return secret, nil
			},
			jwt.WithValidMethods([]string{"HS256"}),
		)

		if err != nil || token == nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		var userID string
		if sub, err := claims.GetSubject(); err == nil && sub != "" {
			userID = sub
		} else if rawSub, ok := claims["sub"]; ok {
			userID = fmt.Sprintf("%v", rawSub)
		}

		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token subject",
			})
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
