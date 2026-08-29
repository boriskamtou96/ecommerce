package services

import (
	"ecommerce/internal/dtos"
	"ecommerce/internal/models"
	"errors"

	"gorm.io/gorm"
)

// Sentinel errors so the HTTP layer can map a cart failure to the right
// status code without matching on error strings.
var (
	ErrCartNotFound      = errors.New("cart not found")
	ErrCartItemNotFound  = errors.New("cart item not found")
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type CartService struct {
	db         *gorm.DB
	cdnBaseURL string
}

func NewCartService(db *gorm.DB, cdnBaseURL string) *CartService {
	return &CartService{db: db, cdnBaseURL: cdnBaseURL}
}

func (s *CartService) GetCart(userID uint) (*dtos.CartResponse, error) {
	var cart models.Cart
	if err := s.db.
		Preload("CartItems.Product.Category").
		Preload("CartItems.Product.Images").
		Where("user_id = ?", userID).
		First(&cart).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCartNotFound
		}
		return nil, err
	}
	return s.convertToCartResponse(&cart), nil
}

func (s *CartService) AddToCart(userID uint, req *dtos.AddToCartRequest) (*dtos.CartResponse, error) {

	// Check if product exists
	var product models.Product
	if err := s.db.First(&product, req.ProductID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	if product.Stock < req.Quantity {
		return nil, ErrInsufficientStock
	}

	// Get or create cart
	var cart models.Cart
	if err := s.db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		cart = models.Cart{UserID: userID}
		if err := s.db.Create(&cart).Error; err != nil {
			return nil, err
		}
	}

	// Check if item already exists in cart
	var cartItem models.CartItem
	if err := s.db.
		Where("cart_id = ? AND product_id = ?", cart.ID, req.ProductID).
		First(&cartItem).Error; err != nil {
		// Create new cart item
		cartItem = models.CartItem{
			CartID:    cart.ID,
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
		}
		if createErr := s.db.Create(&cartItem).Error; createErr != nil {
			return nil, createErr
		}
	} else {
		// Update existing cart item
		cartItem.Quantity += req.Quantity
		if cartItem.Quantity > product.Stock {
			return nil, ErrInsufficientStock
		}
		if saveErr := s.db.Save(&cartItem).Error; saveErr != nil {
			return nil, saveErr
		}
	}

	return s.GetCart(userID)
}

func (s *CartService) UpdateCartItem(userID, itemID uint, req *dtos.UpdateCartItemRequest) (*dtos.CartResponse, error) {
	var cartItem models.CartItem
	if err := s.db.Joins("JOIN carts ON cart_items.cart_id = carts.id").
		Where("cart_items.id = ? AND carts.user_id = ?", itemID, userID).
		First(&cartItem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCartItemNotFound
		}
		return nil, err
	}

	var product models.Product
	if err := s.db.First(&product, cartItem.ProductID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	if product.Stock < req.Quantity {
		return nil, ErrInsufficientStock
	}

	cartItem.Quantity = req.Quantity
	if err := s.db.Save(&cartItem).Error; err != nil {
		return nil, err
	}

	return s.GetCart(userID)
}

func (s *CartService) RemoveFromCart(userID, itemID uint) error {
	result := s.db.Where("id = ? AND cart_id IN (?)", itemID,
		s.db.Select("id").Table("carts").
			Where("user_id = ?", userID)).
		Delete(&models.CartItem{})

	if result.Error != nil {
		return result.Error
	}

	// Nothing matched: either the item does not exist or it belongs to
	// somebody else. Both answer the same way, so an item cannot be probed
	// from another account.
	if result.RowsAffected == 0 {
		return ErrCartItemNotFound
	}

	return nil
}

// ClearCart empties the cart without deleting the cart itself.
func (s *CartService) ClearCart(userID uint) error {
	var cart models.Cart
	if err := s.db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCartNotFound
		}
		return err
	}

	return s.db.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error
}

func (s *CartService) convertToCartResponse(cart *models.Cart) *dtos.CartResponse {
	cartItems := make([]dtos.CartItemResponse, len(cart.CartItems))
	var total float64

	for i := range cart.CartItems {
		subTotal := float64(cart.CartItems[i].Quantity) * cart.CartItems[i].Product.Price
		total += subTotal
		cartItems[i] = dtos.CartItemResponse{
			ID:        cart.CartItems[i].ID,
			Product:   productToResponse(&cart.CartItems[i].Product, s.cdnBaseURL),
			Quantity:  cart.CartItems[i].Quantity,
			Subtotal:  subTotal,
			CreatedAt: cart.CartItems[i].CreatedAt,
			UpdatedAt: cart.CartItems[i].UpdatedAt,
		}
	}

	return &dtos.CartResponse{
		ID:        cart.ID,
		UserID:    cart.UserID,
		CartItems: cartItems,
		Total:     total,
		CreatedAt: cart.CreatedAt,
		UpdatedAt: cart.UpdatedAt,
	}
}
