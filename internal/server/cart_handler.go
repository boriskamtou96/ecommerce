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

// getCart godoc
//
//	@Summary			Get my cart
//	@Description	A user who has never added anything has no cart row yet; that case answers 200 with an empty cart rather than 404.
//	@Tags				cart
//	@Produce			json
//	@Security		BearerAuth
//	@Success			200	{object}	utils.Response{data=dtos.CartResponse}
//	@Failure			401	{object}	utils.Response
//	@Failure			500	{object}	utils.Response
//	@Router			/cart/ [get]
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

// addToCart godoc
//
//	@Summary			Add an item to my cart
//	@Description	Adding a product already in the cart increases its quantity instead of creating a second line. Returns the whole cart.
//	@Tags				cart
//	@Accept			json
//	@Produce			json
//	@Security		BearerAuth
//	@Param				request	body	dtos.AddToCartRequest	true	"Product and quantity"
//	@Success			200	{object}	utils.Response{data=dtos.CartResponse}
//	@Failure			400	{object}	utils.Response
//	@Failure			401	{object}	utils.Response
//	@Failure			404	{object}	utils.Response		"Unknown product"
//	@Failure			409	{object}	utils.Response		"Requested quantity exceeds the stock on hand"
//	@Router			/cart/items [post]
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

// updateCartItem godoc
//
//	@Summary			Set the quantity of a cart item
//	@Description	Replaces the quantity; it does not add to it. Items belonging to another user answer 404.
//	@Tags				cart
//	@Accept			json
//	@Produce			json
//	@Security		BearerAuth
//	@Param				itemId	path	int	true	"Cart item ID"
//	@Param				request	body	dtos.UpdateCartItemRequest	true	"New quantity"
//	@Success			200	{object}	utils.Response{data=dtos.CartResponse}
//	@Failure			400	{object}	utils.Response
//	@Failure			401	{object}	utils.Response
//	@Failure			404	{object}	utils.Response
//	@Failure			409	{object}	utils.Response		"Requested quantity exceeds the stock on hand"
//	@Router			/cart/items/{itemId} [put]
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

// removeFromCart godoc
//
//	@Summary			Remove an item from my cart
//	@Description	Returns the updated cart, or a null payload if the cart no longer exists.
//	@Tags				cart
//	@Produce			json
//	@Security		BearerAuth
//	@Param				itemId	path	int	true	"Cart item ID"
//	@Success			200	{object}	utils.Response{data=dtos.CartResponse}
//	@Failure			400	{object}	utils.Response
//	@Failure			401	{object}	utils.Response
//	@Failure			404	{object}	utils.Response
//	@Router			/cart/items/{itemId} [delete]
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

// clearCart godoc
//
//	@Summary			Empty my cart
//	@Description	Idempotent: emptying an already empty cart still answers 200, so a retry after a network error is harmless.
//	@Tags				cart
//	@Produce			json
//	@Security		BearerAuth
//	@Success			200	{object}	utils.Response
//	@Failure			401	{object}	utils.Response
//	@Failure			500	{object}	utils.Response
//	@Router			/cart/ [delete]
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
