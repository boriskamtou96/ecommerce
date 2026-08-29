package server

import (
	"errors"
	"strconv"

	"ecommerce/internal/services"
	"ecommerce/internal/utils"

	"github.com/gin-gonic/gin"
)

// ==================== ORDER HANDLERS ====================
//
// Every route below is mounted behind the auth middleware, so an order is
// always created for, and looked up by, the authenticated user id. No
// route accepts a user id from the client.

func (s *Server) createOrder(c *gin.Context) {
	userID := c.GetUint("user_id")

	order, err := s.orderService.CreateOrder(userID)
	if err != nil {
		handleOrderError(c, err)
		return
	}

	utils.CreatedResponse(c, "order created successfully", order)
}

func (s *Server) getOrders(c *gin.Context) {
	userID := c.GetUint("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	orders, meta, err := s.orderService.GetUserOrders(userID, page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "failed to fetch orders", err)
		return
	}

	utils.PaginatedSuccessResponse(c, "orders fetched successfully", orders, meta)
}

func (s *Server) getOrder(c *gin.Context) {
	userID := c.GetUint("user_id")

	orderID, ok := orderIDParam(c)
	if !ok {
		return
	}

	order, err := s.orderService.GetUserOrder(userID, orderID)
	if err != nil {
		handleOrderError(c, err)
		return
	}

	utils.SuccessResponse(c, "order fetched successfully", order)
}

func orderIDParam(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "invalid order id", err)
		return 0, false
	}
	return uint(id), true
}

// handleOrderError maps a service error to an HTTP status. An empty or
// missing cart is a 400: the client asked to check out with nothing to
// buy. Out of stock is a 409, as in the cart handler: the request is well
// formed, it just lost the race against the current inventory.
func handleOrderError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrCartNotFound),
		errors.Is(err, services.ErrCartEmpty):
		utils.BadRequestResponse(c, err.Error(), err)
	case errors.Is(err, services.ErrOrderNotFound):
		utils.NotFoundResponse(c, err.Error(), err)
	case errors.Is(err, services.ErrInsufficientStock):
		utils.ConflictResponse(c, err.Error(), err)
	default:
		utils.InternalServerErrorResponse(c, "order operation failed", err)
	}
}
