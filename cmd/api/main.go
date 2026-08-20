package main

import (
	"ecommerce/internal/config"
	"ecommerce/internal/database"
	"ecommerce/internal/logger"

	"github.com/gin-gonic/gin"
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

	log.Info().Msg("Server starting...")

}
