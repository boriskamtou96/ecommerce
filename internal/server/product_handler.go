package server

import (
	"strconv"

	"ecommerce/internal/dtos"
	"ecommerce/internal/utils"

	"github.com/gin-gonic/gin"
)

// ==================== CATEGORY HANDLERS ====================

func (s *Server) createCategory(c *gin.Context) {
	var req dtos.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request", err)
		return
	}

	category, err := s.productService.CreateCategory(&req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "failed to create category", err)
		return
	}

	utils.CreatedResponse(c, "category created successfully", category)
}

func (s *Server) getCategories(c *gin.Context) {
	categories, err := s.productService.GetCategories()
	if err != nil {
		utils.InternalServerErrorResponse(c, "failed to fetch categories", err)
		return
	}

	utils.SuccessResponse(c, "categories fetched successfully", categories)
}

func (s *Server) updateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "invalid category id", err)
		return
	}

	var req dtos.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request", err)
		return
	}

	category, err := s.productService.UpdateCategory(uint(id), &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "failed to update category", err)
		return
	}

	utils.SuccessResponse(c, "category updated successfully", category)
}

func (s *Server) deleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "invalid category id", err)
		return
	}

	if err := s.productService.DeleteCategory(uint(id)); err != nil {
		utils.InternalServerErrorResponse(c, "failed to delete category", err)
		return
	}

	utils.SuccessResponse(c, "category deleted successfully", nil)
}

// ==================== PRODUCT HANDLERS ====================

func (s *Server) createProduct(c *gin.Context) {
	var req dtos.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request", err)
		return
	}

	product, err := s.productService.CreateProduct(&req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "failed to create product", err)
		return
	}

	utils.CreatedResponse(c, "product created successfully", product)
}

func (s *Server) getProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	products, meta, err := s.productService.GetProducts(page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "failed to fetch products", err)
		return
	}

	utils.PaginatedSuccessResponse(c, "products fetched successfully", products, meta)
}

func (s *Server) getProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "invalid product id", err)
		return
	}

	product, err := s.productService.GetProduct(uint(id))
	if err != nil {
		utils.NotFoundResponse(c, "product not found", err)
		return
	}

	utils.SuccessResponse(c, "product fetched successfully", product)
}

func (s *Server) updateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "invalid product id", err)
		return
	}

	var req dtos.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request", err)
		return
	}

	product, err := s.productService.UpdateProduct(uint(id), &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "failed to update product", err)
		return
	}

	utils.SuccessResponse(c, "product updated successfully", product)
}

func (s *Server) deleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "invalid product id", err)
		return
	}

	if err := s.productService.DeleteProduct(uint(id)); err != nil {
		utils.InternalServerErrorResponse(c, "failed to delete product", err)
		return
	}

	utils.SuccessResponse(c, "product deleted successfully", nil)
}

func (s *Server) uploadProductImage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid product ID", err)
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		utils.BadRequestResponse(c, "No file uploaded", err)
		return
	}

	url, err := s.uploadService.UploadProductImage(uint(id), file)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to upload image", err)
		return
	}

	if err := s.productService.AddProductImage(uint(id), url, file.Filename); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to save image record", err)
		return
	}

	utils.SuccessResponse(c, "Image uploaded successfully", map[string]string{"url": url})
}
