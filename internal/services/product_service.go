package services

import (
	"errors"

	"ecommerce/internal/dtos"
	"ecommerce/internal/models"
	"ecommerce/internal/utils"

	"gorm.io/gorm"
)

type ProductService struct {
	db         *gorm.DB
	cdnBaseURL string
}

func NewProductService(db *gorm.DB, cdnBaseURL string) *ProductService {
	return &ProductService{
		db:         db,
		cdnBaseURL: cdnBaseURL,
	}
}

func (s *ProductService) CreateCategory(
	req *dtos.CreateCategoryRequest,
) (*dtos.CategoryResponse, error) {
	category := models.Category{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.db.Create(&category).Error; err != nil {
		return nil, err
	}

	return &dtos.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		IsActive:    category.IsActive,
	}, nil
}

func (s *ProductService) GetCategories() ([]dtos.CategoryResponse, error) {
	var categories []models.Category
	if err := s.db.Where("is_active = ?", true).Find(&categories).Error; err != nil {
		return nil, err
	}

	response := make([]dtos.CategoryResponse, len(categories))

	for i := range categories {
		response[i] = dtos.CategoryResponse{
			ID:          categories[i].ID,
			Name:        categories[i].Name,
			Description: categories[i].Description,
			IsActive:    categories[i].IsActive,
		}
	}
	return response, nil
}

func (s *ProductService) UpdateCategory(
	id uint,
	req *dtos.UpdateCategoryRequest,
) (*dtos.CategoryResponse, error) {
	var category models.Category
	if err := s.db.First(&category, id).Error; err != nil {
		return nil, err
	}

	category.Name = req.Name
	category.Description = req.Description
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := s.db.Save(&category).Error; err != nil {
		return nil, err
	}

	return &dtos.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		IsActive:    category.IsActive,
	}, nil
}

func (s *ProductService) DeleteCategory(id uint) error {
	return s.db.Delete(&models.Category{}, id).Error
}

func (s *ProductService) CreateProduct(req *dtos.CreateProductRequest) (*dtos.ProductResponse, error) {
	product := models.Product{
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		SKU:         req.SKU,
	}

	if err := s.db.Create(&product).Error; err != nil {
		return nil, err
	}

	return s.GetProduct(product.ID)
}

func (s *ProductService) GetProducts(page, limit int) ([]dtos.ProductResponse, *utils.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit
	var products []models.Product
	var total int64

	s.db.Model(&models.Product{}).Where("is_active = ?", true).Count(&total)

	if err := s.db.Preload("Category").Preload("Images").
		Where("is_active = ?", true).
		Offset(offset).Limit(limit).
		Find(&products).Error; err != nil {
		return nil, nil, err
	}

	response := make([]dtos.ProductResponse, len(products))
	for i := range products {
		response[i] = s.convertToProductResponse(&products[i])
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

func (s *ProductService) GetProduct(id uint) (*dtos.ProductResponse, error) {
	var product models.Product
	if err := s.db.Preload("Category").Preload("Images").First(&product, id).Error; err != nil {
		return nil, err
	}

	response := s.convertToProductResponse(&product)
	return &response, nil
}

func (s *ProductService) UpdateProduct(id uint, req *dtos.UpdateProductRequest) (*dtos.ProductResponse, error) {
	var product models.Product
	if err := s.db.First(&product, id).Error; err != nil {
		return nil, err
	}

	product.CategoryID = req.CategoryID
	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Stock = req.Stock
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := s.db.Save(&product).Error; err != nil {
		return nil, err
	}

	return s.GetProduct(id)
}

func (s *ProductService) DeleteProduct(id uint) error {
	return s.db.Delete(&models.Product{}, id).Error
}

// AddProductImage records the storage key (not the public URL) so the CDN
// host can change without a data migration. The created row is returned so
// the caller can hand its id back to the client, which would otherwise have
// to refetch the product just to be able to delete what it just uploaded.
func (s *ProductService) AddProductImage(
	productID uint,
	key, altText string,
) (*models.ProductImage, error) {
	var count int64
	s.db.Model(&models.ProductImage{}).Where("product_id = ?", productID).Count(&count)

	image := models.ProductImage{
		ProductID: productID,
		URL:       key,
		AltText:   altText,
		IsPrimary: count == 0, // First image is primary
	}

	if err := s.db.Create(&image).Error; err != nil {
		return nil, err
	}

	return &image, nil
}

func (s *ProductService) GetProductImage(productID, imageID uint) (*models.ProductImage, error) {
	var image models.ProductImage
	if err := s.db.Where("product_id = ? AND id = ?", productID, imageID).
		First(&image).Error; err != nil {
		return nil, err
	}
	return &image, nil
}

// DeleteProductImage removes the row and promotes another image to primary
// when the deleted one was the primary.
func (s *ProductService) DeleteProductImage(image *models.ProductImage) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Delete(&models.ProductImage{}, image.ID).Error; err != nil {
			return err
		}

		if !image.IsPrimary {
			return nil
		}

		var next models.ProductImage
		err := tx.Where("product_id = ?", image.ProductID).Order("id asc").First(&next).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}

		return tx.Model(&next).Update("is_primary", true).Error
	})
}

func (s *ProductService) convertToProductResponse(product *models.Product) dtos.ProductResponse {
	return productToResponse(product, s.cdnBaseURL)
}

// productToResponse is shared by every service that embeds a product in
// its payload (catalogue, cart, orders) so that image URLs are built the
// same way everywhere.
func productToResponse(product *models.Product, cdnBaseURL string) dtos.ProductResponse {
	images := make([]dtos.ProductImageResponse, len(product.Images))
	for i := range product.Images {
		images[i] = dtos.ProductImageResponse{
			ID:        product.Images[i].ID,
			URL:       utils.CDNURL(cdnBaseURL, product.Images[i].URL),
			AltText:   product.Images[i].AltText,
			IsPrimary: product.Images[i].IsPrimary,
		}
	}
	return dtos.ProductResponse{
		ID:          product.ID,
		CategoryID:  product.CategoryID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		SKU:         product.SKU,
		IsActive:    product.IsActive,
		Category: dtos.CategoryResponse{
			ID:          product.Category.ID,
			Name:        product.Category.Name,
			Description: product.Category.Description,
			IsActive:    product.Category.IsActive,
		},
		Images: images,
	}
}
