package server

import (
	"ecommerce/internal/config"
	"ecommerce/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Server struct {
	db             *gorm.DB
	config         *config.Config
	logger         zerolog.Logger
	authService    *services.AuthService
	productService *services.ProductService
	userService    *services.UserService
	uploadService  *services.UploadService
}

func New(
	db *gorm.DB,
	config *config.Config,
	logger zerolog.Logger,
	authService *services.AuthService,
	productService *services.ProductService,
	userService *services.UserService,
	uploadService *services.UploadService,
) *Server {
	return &Server{
		db:             db,
		config:         config,
		logger:         logger,
		authService:    authService,
		productService: productService,
		userService:    userService,
		uploadService:  uploadService,
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
				products.POST("/:id/images", s.adminMiddleware(), s.uploadProductImage)
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
