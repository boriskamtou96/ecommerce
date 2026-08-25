package server

import (
	"ecommerce/internal/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Server struct {
	db     *gorm.DB
	config *config.Config
	logger zerolog.Logger
}

func New(db *gorm.DB, config *config.Config, logger zerolog.Logger) *Server {
	return &Server{
		db:     db,
		config: config,
		logger: logger,
	}
}

func (s *Server) SetupRoutes() *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(s.corsMiddleware())

	router.GET("/health", healthCheck)

	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", s.register)
			auth.POST("/login", s.login)
			auth.POST("/refresh", s.refreshToken)
			auth.POST("/logout", s.logout)
		}

		api.GET("/categories", s.getCategories)
		api.GET("/products", s.getProducts)
		api.GET("/products/:id", s.getProduct)

		protectedRoutes := api.Group("/")
		protectedRoutes.Use(s.authMiddleware())
		{
			users := protectedRoutes.Group("/users")
			{
				users.GET("/profile", s.getProfile)
				users.PUT("/profile", s.updateProfile)
			}

			products := protectedRoutes.Group("/products")
			{
				products.POST("/", s.adminMiddleware(), s.createProduct)
				products.PUT("/:id", s.adminMiddleware(), s.updateProduct)
				products.DELETE("/:id", s.adminMiddleware(), s.deleteProduct)
			}

			categories := protectedRoutes.Group("/categories")
			{
				categories.POST("/", s.adminMiddleware(), s.createCategory)
				categories.PUT("/:id", s.adminMiddleware(), s.updateCategory)
				categories.DELETE("/:id", s.adminMiddleware(), s.deleteCategory)
			}
		}
	}

	return router
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "Ok"})
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
