package server

import (
	"ecommerce/internal/models"
	"ecommerce/internal/utils"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.UnauthorizedResponse(c, "you need to provide authorization", errors.New("you need to provide authorization"))
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) < 2 || tokenParts[0] != "Bearer" {
			utils.UnauthorizedResponse(c, "invalid authorization header format", errors.New("invalid authorization header format"))
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(tokenParts[1], s.config.JWT.Secret)
		if err != nil {
			utils.UnauthorizedResponse(c, "invalid token", errors.New("invalid token"))
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

func (s *Server) adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			utils.ForbiddenResponse(c, "forbidden")
			c.Abort()
			return
		}

		if role != string(models.UserRoleAdmin) {
			utils.ForbiddenResponse(c, "forbidden")
			c.Abort()
			return
		}
		c.Next()
	}
}

func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Methods", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
