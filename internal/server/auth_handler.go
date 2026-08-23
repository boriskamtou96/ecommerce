package server

import (
	"ecommerce/internal/dtos"
	"ecommerce/internal/services"
	"ecommerce/internal/utils"
	"log"

	"github.com/gin-gonic/gin"
)

func (s *Server) register(c *gin.Context) {
	var req dtos.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request", err)
		return
	}

	authService := services.NewAuthService(s.db, s.config)
	response, err := authService.Register(&req)
	if err != nil {
		log.Printf("register failed: %+v", err)
		utils.BadRequestResponse(c, "registration failed", err)
		return
	}

	utils.CreatedResponse(c, "user register successfully", response)
}

func (s *Server) login(c *gin.Context) {
	var req dtos.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request", err)
		return
	}

	authService := services.NewAuthService(s.db, s.config)
	response, err := authService.Login(&req)
	if err != nil {
		utils.BadRequestResponse(c, "login failed", err)
		return
	}

	utils.CreatedResponse(c, "login successfully", response)
}

func (s *Server) refreshToken(c *gin.Context) {
	var req dtos.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request", err)
		return
	}

	authService := services.NewAuthService(s.db, s.config)
	response, err := authService.RefreshToken(&req)
	if err != nil {
		utils.BadRequestResponse(c, "token refresh failed", err)
		return
	}

	utils.CreatedResponse(c, "token refreshed successfully", response)
}

func (s *Server) logout(c *gin.Context) {
	var req dtos.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request", err)
		return
	}

	authService := services.NewAuthService(s.db, s.config)
	err := authService.Logout(req.RefreshToken)
	if err != nil {
		utils.InternalServerErrorResponse(c, "token refresh failed", err)
		return
	}

	utils.SuccessResponse(c, "Logout successful", nil)
}
