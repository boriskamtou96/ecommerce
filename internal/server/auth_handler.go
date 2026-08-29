package server

import (
	"ecommerce/internal/dtos"
	"ecommerce/internal/utils"
	"log"

	"github.com/gin-gonic/gin"
)

// register godoc
//
//	@Summary			Register a new account
//	@Description	Creates a customer account and returns it with a fresh token pair.
//	@Tags				auth
//	@Accept			json
//	@Produce			json
//	@Param				request	body	dtos.RegisterRequest	true	"Account details (password: 8 characters minimum)"
//	@Success			201	{object}	utils.Response{data=dtos.AuthResponse}
//	@Failure			400	{object}	utils.Response		"Malformed payload, or the email is already taken"
//	@Router			/auth/register [post]
func (s *Server) register(c *gin.Context) {
	var req dtos.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request", err)
		return
	}

	response, err := s.authService.Register(&req)
	if err != nil {
		log.Printf("register failed: %+v", err)
		utils.BadRequestResponse(c, "registration failed", err)
		return
	}

	utils.CreatedResponse(c, "user register successfully", response)
}

// login godoc
//
//	@Summary			Log in
//	@Description	Exchanges credentials for an access token and a refresh token.
//	@Tags				auth
//	@Accept			json
//	@Produce			json
//	@Param				request	body	dtos.LoginRequest	true	"Credentials"
//	@Success			201	{object}	utils.Response{data=dtos.AuthResponse}
//	@Failure			400	{object}	utils.Response		"Unknown email or wrong password"
//	@Router			/auth/login [post]
func (s *Server) login(c *gin.Context) {
	var req dtos.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request", err)
		return
	}

	response, err := s.authService.Login(&req)
	if err != nil {
		utils.BadRequestResponse(c, "login failed", err)
		return
	}

	utils.CreatedResponse(c, "login successfully", response)
}

// refreshToken godoc
//
//	@Summary			Refresh the access token
//	@Description	Issues a new token pair from a valid refresh token.
//	@Tags				auth
//	@Accept			json
//	@Produce			json
//	@Param				request	body	dtos.RefreshTokenRequest	true	"Refresh token"
//	@Success			201	{object}	utils.Response{data=dtos.AuthResponse}
//	@Failure			400	{object}	utils.Response		"Expired, revoked or unknown refresh token"
//	@Router			/auth/refresh [post]
func (s *Server) refreshToken(c *gin.Context) {
	var req dtos.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request", err)
		return
	}

	response, err := s.authService.RefreshToken(&req)
	if err != nil {
		utils.BadRequestResponse(c, "token refresh failed", err)
		return
	}

	utils.CreatedResponse(c, "token refreshed successfully", response)
}

// logout godoc
//
//	@Summary			Log out
//	@Description	Revokes the refresh token. The access token stays valid until it expires on its own.
//	@Tags				auth
//	@Accept			json
//	@Produce			json
//	@Param				request	body	dtos.RefreshTokenRequest	true	"Refresh token to revoke"
//	@Success			200	{object}	utils.Response
//	@Failure			400	{object}	utils.Response
//	@Failure			500	{object}	utils.Response
//	@Router			/auth/logout [post]
func (s *Server) logout(c *gin.Context) {
	var req dtos.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request", err)
		return
	}

	err := s.authService.Logout(req.RefreshToken)
	if err != nil {
		utils.InternalServerErrorResponse(c, "token refresh failed", err)
		return
	}

	utils.SuccessResponse(c, "Logout successful", nil)
}

// getProfile godoc
//
//	@Summary			Get my profile
//	@Tags				users
//	@Produce			json
//	@Security		BearerAuth
//	@Success			200	{object}	utils.Response{data=dtos.UserResponse}
//	@Failure			401	{object}	utils.Response
//	@Failure			404	{object}	utils.Response
//	@Router			/users/profile [get]
func (s *Server) getProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	profile, err := s.userService.GetProfile(userID)
	if err != nil {
		utils.NotFoundResponse(c, "user not found", err)
		return
	}

	utils.SuccessResponse(c, "profile retrieved successfully", profile)
}

// updateProfile godoc
//
//	@Summary			Update my profile
//	@Description	Updates the caller's own profile. Email and role cannot be changed here.
//	@Tags				users
//	@Accept			json
//	@Produce			json
//	@Security		BearerAuth
//	@Param				request	body	dtos.UpdateProfileRequest	true	"New profile values"
//	@Success			200	{object}	utils.Response{data=dtos.UserResponse}
//	@Failure			400	{object}	utils.Response
//	@Failure			401	{object}	utils.Response
//	@Failure			500	{object}	utils.Response
//	@Router			/users/profile [put]
func (s *Server) updateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req dtos.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request data", err)
		return
	}

	profile, err := s.userService.UpdateProfile(userID, &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "failed to update profile", err)
		return
	}

	utils.SuccessResponse(c, "profile updated successfully", profile)
}
