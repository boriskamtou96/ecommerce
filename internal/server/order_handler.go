package server

import (
	"errors"
	"strconv"

	// swag resolves the types named in the annotations below through the
	// imports of the file they sit in, so dtos must be imported here even
	// though this handler only passes the service's DTOs through.
	_ "ecommerce/internal/dtos"
	"ecommerce/internal/services"
	"ecommerce/internal/utils"

	"github.com/gin-gonic/gin"
)

// ==================== ORDER HANDLERS ====================
//
// Every route below is mounted behind the auth middleware, so an order is
// always created for, and looked up by, the authenticated user id. No
// route accepts a user id from the client.

// createOrder godoc
//
//	@Summary			Check out: turn my cart into an order
//	@Description	Takes no payload: the authenticated user's cart is the input. In one transaction the stock of every product is decremented, a pending order is created and the cart is emptied. Any failure rolls the whole thing back, so a 409 leaves every stock untouched.
//	@Tags				orders
//	@Produce			json
//	@Security		BearerAuth
//	@Success			201	{object}	utils.Response{data=dtos.OrderResponse}
//	@Failure			400	{object}	utils.Response		"Cart is empty or does not exist"
//	@Failure			401	{object}	utils.Response
//	@Failure			409	{object}	utils.Response		"A product ran out of stock before the order could be placed"
//	@Failure			500	{object}	utils.Response
//	@Router			/orders/ [post]
func (s *Server) createOrder(c *gin.Context) {
	userID := c.GetUint("user_id")

	order, err := s.orderService.CreateOrder(userID)
	if err != nil {
		handleOrderError(c, err)
		return
	}

	utils.CreatedResponse(c, "order created successfully", order)
}

// getOrders godoc
//
//	@Summary			List my orders
//	@Description	Paginated, newest first, restricted to the authenticated user. A user with no orders gets an empty list, not a 404.
//	@Tags				orders
//	@Produce			json
//	@Security		BearerAuth
//	@Param				page	query	int	false	"Page number, 1 based"	default(1)
//	@Param				limit	query	int	false	"Items per page"	default(10)
//	@Success			200	{object}	utils.PaginatedResponse{data=[]dtos.OrderResponse}
//	@Failure			401	{object}	utils.Response
//	@Failure			500	{object}	utils.Response
//	@Router			/orders/ [get]
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

// getOrder godoc
//
//	@Summary			Get one of my orders
//	@Description	The ownership check is part of the query, so an order belonging to somebody else answers 404 rather than 403: order IDs cannot be probed.
//	@Tags				orders
//	@Produce			json
//	@Security		BearerAuth
//	@Param				id	path	int	true	"Order ID"
//	@Success			200	{object}	utils.Response{data=dtos.OrderResponse}
//	@Failure			400	{object}	utils.Response		"Non numeric id"
//	@Failure			401	{object}	utils.Response
//	@Failure			404	{object}	utils.Response		"Unknown order, or not yours"
//	@Router			/orders/{id} [get]
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
