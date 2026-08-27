package server

import (
	"errors"
	"strconv"

	"ecommerce/internal/dtos"
	"ecommerce/internal/services"
	"ecommerce/internal/utils"

	"github.com/gin-gonic/gin"
)

// ==================== CART HANDLERS ====================
//
// Every route below is mounted behind the auth middleware, so the cart is
// always resolved from the authenticated user id rather than from a
// client supplied identifier. A user can therefore never reach somebody
// else's cart.

func (s *Server) getCart(c *gin.Context) {
	userID := c.GetUint("user_id")

	cart, err := s.cartService.GetCart(userID)
	if err != nil {
		// A user who never added anything has no cart row yet. That is an
		// empty cart, not a 404, otherwise every front end has to special
		// case a brand new account.
		if errors.Is(err, services.ErrCartNotFound) {
			utils.SuccessResponse(c, "cart retrieved successfully", &dtos.CartResponse{
				UserID:    userID,
				CartItems: []dtos.CartItemResponse{},
				Total:     0,
			})
			return
		}
		handleCartError(c, err)
		return
	}

	utils.SuccessResponse(c, "cart retrieved successfully", cart)
}

func (s *Server) addToCart(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req dtos.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request", err)
		return
	}

	cart, err := s.cartService.AddToCart(userID, &req)
	if err != nil {
		handleCartError(c, err)
		return
	}

	utils.SuccessResponse(c, "item added to cart successfully", cart)
}

func (s *Server) updateCartItem(c *gin.Context) {
	userID := c.GetUint("user_id")

	itemID, ok := cartItemIDParam(c)
	if !ok {
		return
	}

	var req dtos.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request", err)
		return
	}

	cart, err := s.cartService.UpdateCartItem(userID, itemID, &req)
	if err != nil {
		handleCartError(c, err)
		return
	}

	utils.SuccessResponse(c, "cart item updated successfully", cart)
}

func (s *Server) removeFromCart(c *gin.Context) {
	userID := c.GetUint("user_id")

	itemID, ok := cartItemIDParam(c)
	if !ok {
		return
	}

	if err := s.cartService.RemoveFromCart(userID, itemID); err != nil {
		handleCartError(c, err)
		return
	}

	cart, err := s.cartService.GetCart(userID)
	if err != nil {
		if errors.Is(err, services.ErrCartNotFound) {
			utils.SuccessResponse(c, "item removed from cart successfully", nil)
			return
		}
		handleCartError(c, err)
		return
	}

	utils.SuccessResponse(c, "item removed from cart successfully", cart)
}

func (s *Server) clearCart(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Emptying an already empty cart is not a failure: the endpoint is
	// idempotent so a retry after a network error is harmless.
	if err := s.cartService.ClearCart(userID); err != nil &&
		!errors.Is(err, services.ErrCartNotFound) {
		handleCartError(c, err)
		return
	}

	utils.SuccessResponse(c, "cart cleared successfully", nil)
}

func cartItemIDParam(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("itemId"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "invalid cart item id", err)
		return 0, false
	}
	return uint(id), true
}

// handleCartError maps a service error to an HTTP status. Out of stock is
// a 409 rather than a 400: the request is valid, it just conflicts with
// the current state of the inventory.
func handleCartError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrCartNotFound),
		errors.Is(err, services.ErrCartItemNotFound),
		errors.Is(err, services.ErrProductNotFound):
		utils.NotFoundResponse(c, err.Error(), err)
	case errors.Is(err, services.ErrInsufficientStock):
		utils.ConflictResponse(c, err.Error(), err)
	default:
		utils.InternalServerErrorResponse(c, "cart operation failed", err)
	}
}
