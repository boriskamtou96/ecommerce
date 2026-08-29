package main

import (
	"context"
	"ecommerce/internal/config"
	"ecommerce/internal/database"
	"ecommerce/internal/interfaces"
	"ecommerce/internal/logger"
	"ecommerce/internal/providers"
	"ecommerce/internal/server"
	"ecommerce/internal/services"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	requestTimeOut = 10
	cancelTimeout  = 15
)

func main() {
	log := logger.New()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	mainDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get database connection")
	}

	defer mainDB.Close()

	gin.SetMode(cfg.Server.GinMode)

	authService := services.NewAuthService(db, cfg)
	productService := services.NewProductService(db, cfg.CDN.BaseURL)
	userService := services.NewUserService(db)
	cartService := services.NewCartService(db, cfg.CDN.BaseURL)
	orderService := services.NewOrderService(db, cfg.CDN.BaseURL)

	var uploadProvider interfaces.UploadProvider
	if cfg.Upload.UploadProvider == "s3" {
		s3Provider, s3Err := providers.NewS3Provider(cfg)
		if s3Err != nil {
			log.Fatal().Err(s3Err).Msg("Failed to create S3 provider")
		}
		uploadProvider = s3Provider
	} else {
		uploadProvider = providers.NewLocalUploadProvider(cfg.Upload.Path)
	}

	uploadService := services.NewUploadService(uploadProvider, cfg.Upload.MaxFileSize)

	srv := server.New(
		db,
		cfg,
		log,
		authService,
		productService,
		userService,
		uploadService,
		cartService,
		orderService,
	)

	router := srv.SetupRoutes()

	// Cap what Gin buffers in memory for multipart requests; the per-file
	// limit itself is enforced in the upload service.
	router.MaxMultipartMemory = cfg.Upload.MaxFileSize

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  requestTimeOut * time.Second,
		WriteTimeout: requestTimeOut * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.Server.Port).Msg("Server running on")
		if listenServerErr := httpServer.ListenAndServe(); listenServerErr != nil &&
			!errors.Is(listenServerErr, http.ErrServerClosed) {
			log.Fatal().Err(listenServerErr).Msg("failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), cancelTimeout*time.Second)
	defer cancel()

	if shutDownErr := httpServer.Shutdown(ctx); shutDownErr != nil {
		log.Error().Err(shutDownErr).Msg("failed to shutdown http server")
		return
	}

	log.Info().Msg("shutting down database")
}
