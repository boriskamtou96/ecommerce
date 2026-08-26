package main

import (
	"context"
	"ecommerce/internal/config"
	"ecommerce/internal/database"
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
	productService := services.NewProductService(db)
	userService := services.NewUserService(db)
	uploadService := services.NewUploadService(providers.NewLocalUploadProvider(cfg.Upload.Path))

	srv := server.New(db, cfg, log, authService, productService, userService, uploadService)

	router := srv.SetupRoutes()

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
