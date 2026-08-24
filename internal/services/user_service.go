package services

import (
	"ecommerce/internal/config"
	"ecommerce/internal/dtos"
	"ecommerce/internal/models"

	"gorm.io/gorm"
)

type UserService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetProfile(userID uint) (*dtos.UserResponse, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	return &dtos.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Phone:     user.Phone,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      string(user.Role),
		IsActive:  user.IsActive,
	}, nil
}

func (s *UserService) UpdateProfile(userID uint, req *dtos.UpdateProfileRequest) (*dtos.UserResponse, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Phone = req.Phone

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	return s.GetProfile(userID)
}
