package server

import (
	"errors"
	"net/http"
	"strconv"

	"ecommerce/internal/dtos"
	"ecommerce/internal/services"
	"ecommerce/internal/utils"

	"github.com/gin-gonic/gin"
)

// ==================== CATEGORY HANDLERS ====================

// createCategory godoc
//
//	@Summary			Create a category
//	@Tags				categories
//	@Accept			json
//	@Produce			json
//	@Security		BearerAuth
//	@Param				request	body	dtos.CreateCategoryRequest	true	"Category to create"
//	@Success			201	{object}	utils.Response{data=dtos.CategoryResponse}
//	@Failure			400	{object}	utils.Response
//	@Failure			401	{object}	utils.Response
//	@Failure			403	{object}	utils.Response		"Admin role required"
//	@Failure			500	{object}	utils.Response
//	@Router			/categories/ [post]
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

// getCategories godoc
//
//	@Summary			List active categories
//	@Description	Public endpoint. Inactive categories are never returned.
//	@Tags				categories
//	@Produce			json
//	@Success			200	{object}	utils.Response{data=[]dtos.CategoryResponse}
//	@Failure			500	{object}	utils.Response
//	@Router			/categories [get]
func (s *Server) getCategories(c *gin.Context) {
	categories, err := s.productService.GetCategories()
	if err != nil {
		utils.InternalServerErrorResponse(c, "failed to fetch categories", err)
		return
	}

	utils.SuccessResponse(c, "categories fetched successfully", categories)
}

// updateCategory godoc
//
//	@Summary			Update a category
//	@Tags				categories
//	@Accept			json
//	@Produce			json
//	@Security		BearerAuth
//	@Param				id	path	int	true	"Category ID"
//	@Param				request	body	dtos.UpdateCategoryRequest	true	"New values"
//	@Success			200	{object}	utils.Response{data=dtos.CategoryResponse}
//	@Failure			400	{object}	utils.Response
//	@Failure			401	{object}	utils.Response
//	@Failure			403	{object}	utils.Response
//	@Failure			500	{object}	utils.Response
//	@Router			/categories/{id} [put]
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

// deleteCategory godoc
//
//	@Summary			Delete a category
//	@Description	Soft delete: the row is kept with a deleted_at timestamp.
//	@Tags				categories
//	@Produce			json
//	@Security		BearerAuth
//	@Param				id	path	int	true	"Category ID"
//	@Success			200	{object}	utils.Response
//	@Failure			400	{object}	utils.Response
//	@Failure			401	{object}	utils.Response
//	@Failure			403	{object}	utils.Response
//	@Failure			500	{object}	utils.Response
//	@Router			/categories/{id} [delete]
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

// createProduct godoc
//
//	@Summary			Create a product
//	@Tags				products
//	@Accept			json
//	@Produce			json
//	@Security		BearerAuth
//	@Param				request	body	dtos.CreateProductRequest	true	"Product to create (SKU must be unique)"
//	@Success			201	{object}	utils.Response{data=dtos.ProductResponse}
//	@Failure			400	{object}	utils.Response
//	@Failure			401	{object}	utils.Response
//	@Failure			403	{object}	utils.Response
//	@Failure			500	{object}	utils.Response
//	@Router			/products/ [post]
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

// getProducts godoc
//
//	@Summary			List active products
//	@Description	Public and paginated. An out of range page returns an empty list, not an error.
//	@Tags				products
//	@Produce			json
//	@Param				page	query	int	false	"Page number, 1 based"	default(1)
//	@Param				limit	query	int	false	"Items per page"	default(10)
//	@Success			200	{object}	utils.PaginatedResponse{data=[]dtos.ProductResponse}
//	@Failure			500	{object}	utils.Response
//	@Router			/products [get]
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

// getProduct godoc
//
//	@Summary			Get one product
//	@Description	Public. Includes the category and every image of the product.
//	@Tags				products
//	@Produce			json
//	@Param				id	path	int	true	"Product ID"
//	@Success			200	{object}	utils.Response{data=dtos.ProductResponse}
//	@Failure			400	{object}	utils.Response		"Non numeric id"
//	@Failure			404	{object}	utils.Response
//	@Router			/products/{id} [get]
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

// updateProduct godoc
//
//	@Summary			Update a product
//	@Description	Full replacement: every field of the request is written, so send the current values for the ones you do not change.
//	@Tags				products
//	@Accept			json
//	@Produce			json
//	@Security		BearerAuth
//	@Param				id	path	int	true	"Product ID"
//	@Param				request	body	dtos.UpdateProductRequest	true	"New values"
//	@Success			200	{object}	utils.Response{data=dtos.ProductResponse}
//	@Failure			400	{object}	utils.Response
//	@Failure			401	{object}	utils.Response
//	@Failure			403	{object}	utils.Response
//	@Failure			500	{object}	utils.Response
//	@Router			/products/{id} [put]
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

// deleteProduct godoc
//
//	@Summary			Delete a product
//	@Description	Soft delete. Existing order items keep pointing at the product.
//	@Tags				products
//	@Produce			json
//	@Security		BearerAuth
//	@Param				id	path	int	true	"Product ID"
//	@Success			200	{object}	utils.Response
//	@Failure			400	{object}	utils.Response
//	@Failure			401	{object}	utils.Response
//	@Failure			403	{object}	utils.Response
//	@Failure			500	{object}	utils.Response
//	@Router			/products/{id} [delete]
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

// uploadProductImage godoc
//
//	@Summary			Upload a product image
//	@Description	Multipart upload. The object is stored by the configured provider (local or S3) and the returned URL is already prefixed with the CDN base URL.
//	@Tags				products
//	@Accept			multipart/form-data
//	@Produce			json
//	@Security		BearerAuth
//	@Param				id	path	int	true	"Product ID"
//	@Param				image	formData	file	true	"Image file"
//	@Success			201	{object}	utils.Response{data=dtos.ProductImageResponse}
//	@Failure			400	{object}	utils.Response
//	@Failure			401	{object}	utils.Response
//	@Failure			403	{object}	utils.Response
//	@Failure			404	{object}	utils.Response
//	@Failure			413	{object}	utils.Response		"File larger than UPLOAD_MAX_FILE_SIZE"
//	@Failure			500	{object}	utils.Response
//	@Router			/products/{id}/images [post]
func (s *Server) uploadProductImage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid product ID", err)
		return
	}

	if _, err = s.productService.GetProduct(uint(id)); err != nil {
		utils.NotFoundResponse(c, "product not found", err)
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		utils.BadRequestResponse(c, "No file uploaded", err)
		return
	}

	key, err := s.uploadService.UploadProductImage(c.Request.Context(), uint(id), file)
	if err != nil {
		if errors.Is(err, services.ErrFileTooLarge) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"success": false,
				"message": "File too large",
				"error":   err.Error(),
			})
			return
		}
		utils.BadRequestResponse(c, "Failed to upload image", err)
		return
	}

	image, err := s.productService.AddProductImage(uint(id), key, file.Filename)
	if err != nil {
		// Keep storage and database consistent: drop the orphan object.
		if delErr := s.uploadService.DeleteFile(c.Request.Context(), key); delErr != nil {
			s.logger.Error().Err(delErr).Str("key", key).Msg("failed to clean up orphaned upload")
		}
		utils.InternalServerErrorResponse(c, "Failed to save image record", err)
		return
	}

	utils.CreatedResponse(c, "Image uploaded successfully", dtos.ProductImageResponse{
		ID:        image.ID,
		URL:       utils.CDNURL(s.config.CDN.BaseURL, key),
		AltText:   image.AltText,
		IsPrimary: image.IsPrimary,
	})
}

// deleteProductImage godoc
//
//	@Summary			Delete a product image
//	@Description	Removes the database row first. A storage failure afterwards is logged, not returned, so the call stays successful.
//	@Tags				products
//	@Produce			json
//	@Security		BearerAuth
//	@Param				id	path	int	true	"Product ID"
//	@Param				imageId	path	int	true	"Image ID"
//	@Success			200	{object}	utils.Response
//	@Failure			400	{object}	utils.Response
//	@Failure			401	{object}	utils.Response
//	@Failure			403	{object}	utils.Response
//	@Failure			404	{object}	utils.Response
//	@Failure			500	{object}	utils.Response
//	@Router			/products/{id}/images/{imageId} [delete]
func (s *Server) deleteProductImage(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid product ID", err)
		return
	}

	imageID, err := strconv.ParseUint(c.Param("imageId"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid image ID", err)
		return
	}

	image, err := s.productService.GetProductImage(uint(productID), uint(imageID))
	if err != nil {
		utils.NotFoundResponse(c, "image not found", err)
		return
	}

	if err = s.productService.DeleteProductImage(image); err != nil {
		utils.InternalServerErrorResponse(c, "failed to delete image record", err)
		return
	}

	// The row is gone, so a storage failure must not fail the request:
	// log it and let a cleanup job pick up the leftover object.
	if err = s.uploadService.DeleteFile(c.Request.Context(), image.URL); err != nil {
		s.logger.Error().Err(err).Str("key", image.URL).Msg("failed to delete object from storage")
	}

	utils.SuccessResponse(c, "image deleted successfully", nil)
}
