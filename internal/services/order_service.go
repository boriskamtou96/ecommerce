package services

import (
	"ecommerce/internal/dtos"
	"ecommerce/internal/models"
	"ecommerce/internal/utils"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// ErrCartNotFound, ErrInsufficientStock and friends already live in
// cart_service.go — same package, so they are reused here rather than
// redeclared.
var (
	ErrCartEmpty     = errors.New("cart is empty")
	ErrOrderNotFound = errors.New("order not found")
)

type OrderService struct {
	db         *gorm.DB
	cdnBaseURL string
}

func NewOrderService(db *gorm.DB, cdnBaseURL string) *OrderService {
	return &OrderService{db: db, cdnBaseURL: cdnBaseURL}
}

// CreateOrder returns order create by user with id userID
func (s *OrderService) CreateOrder(userID uint) (*dtos.OrderResponse, error) {
	var orderResponse *dtos.OrderResponse

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var cart models.Cart
		if err := tx.Preload("CartItems.Product").
			Where("user_id = ?", userID).
			First(&cart).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCartNotFound
			}
			return err
		}

		if len(cart.CartItems) == 0 {
			return ErrCartEmpty
		}

		var totalAmount float64
		orderItems := make([]models.OrderItem, 0, len(cart.CartItems))

		for i := range cart.CartItems {
			cartItem := &cart.CartItems[i]

			// Atomic decrement: the "enough stock?" test is evaluated by the
			// database inside the UPDATE, so two concurrent checkouts cannot
			// both pass it and oversell the product.
			res := tx.Model(&models.Product{}).
				Where("id = ? AND stock >= ?", cartItem.ProductID, cartItem.Quantity).
				Update("stock", gorm.Expr("stock - ?", cartItem.Quantity))
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return fmt.Errorf("%w: %s", ErrInsufficientStock, cartItem.Product.Name)
			}

			totalAmount += float64(cartItem.Quantity) * cartItem.Product.Price

			orderItems = append(orderItems, models.OrderItem{
				ProductID: cartItem.ProductID,
				Quantity:  cartItem.Quantity,
				Price:     cartItem.Product.Price,
			})
		}

		// One order for the whole cart, created after the loop.
		order := models.Order{
			UserID:      userID,
			Status:      models.OrderStatusPending,
			TotalAmount: totalAmount,
			OrderItems:  orderItems,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		if err := tx.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
			return err
		}

		response, err := s.getOrderResponse(tx, order.ID)
		if err != nil {
			return err
		}
		orderResponse = response

		return nil
	})

	if err != nil {
		return nil, err
	}

	return orderResponse, nil
}

// GetUserOrders returns one page of the user's orders, newest first.
func (s *OrderService) GetUserOrders(
	userID uint,
	page, limit int,
) ([]dtos.OrderResponse, *utils.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	var total int64
	if err := s.db.Model(&models.Order{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, nil, err
	}

	var orders []models.Order
	if err := s.db.
		Preload("OrderItems.Product.Category").
		Preload("OrderItems.Product.Images").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&orders).Error; err != nil {
		return nil, nil, err
	}

	// No rows is an empty page, not an error: a user with no orders yet is a
	// normal state, and the handler should answer 200 with [].
	response := make([]dtos.OrderResponse, len(orders))
	for i := range orders {
		response[i] = s.convertToOrderResponse(&orders[i])
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	meta := &utils.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return response, meta, nil
}

// GetUserOrder returns a single order, scoped to its owner. The user_id
// condition is part of the query rather than a check on the loaded row, so
// another user's order simply does not match and cannot be probed by ID.
func (s *OrderService) GetUserOrder(userID, orderID uint) (*dtos.OrderResponse, error) {
	var order models.Order
	if err := s.db.
		Preload("OrderItems.Product.Category").
		Preload("OrderItems.Product.Images").
		Where("id = ? AND user_id = ?", orderID, userID).
		First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	response := s.convertToOrderResponse(&order)

	return &response, nil
}

func (s *OrderService) getOrderResponse(tx *gorm.DB, orderID uint) (*dtos.OrderResponse, error) {
	var order models.Order
	if err := tx.
		Preload("OrderItems.Product.Category").
		Preload("OrderItems.Product.Images").
		First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	response := s.convertToOrderResponse(&order)

	return &response, nil
}

func (s *OrderService) convertToOrderResponse(order *models.Order) dtos.OrderResponse {
	orderItems := make([]dtos.OrderItemResponse, len(order.OrderItems))

	for i := range order.OrderItems {
		orderItems[i] = dtos.OrderItemResponse{
			ID:        order.OrderItems[i].ID,
			Product:   productToResponse(&order.OrderItems[i].Product, s.cdnBaseURL),
			Quantity:  order.OrderItems[i].Quantity,
			Price:     order.OrderItems[i].Price,
			CreatedAt: order.OrderItems[i].CreatedAt,
		}
	}

	return dtos.OrderResponse{
		ID:          order.ID,
		UserID:      order.UserID,
		Status:      string(order.Status),
		TotalAmount: order.TotalAmount,
		OrderItems:  orderItems,
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}
}
