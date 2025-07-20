package repositories

import (
	"errors"
	"fmt"
	"sso-service/app/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	GetUserByUsername(username string) (models.User, error)
	GetUserByID(id uuid.UUID) (models.User, error)
	GetAllUsers() ([]models.User, error)
	CreateUser(user models.User) (models.User, error)
	UpdateUser(user models.User) (models.User, error)
	DeleteUser(id uuid.UUID) error
}

// UserRepositoryImpl is the implementation of UserRepository
type UserRepositoryImpl struct {
	db *gorm.DB // Diubah ke *gorm.DB
}

// NewUserRepository creates a new instance of UserRepositoryImpl
func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

// GetUserByUsername retrieves a user by their username from the database using GORM
func (r *UserRepositoryImpl) GetUserByUsername(username string) (models.User, error) {
	var user models.User
	result := r.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.User{}, fmt.Errorf("user not found: %w", result.Error)
		}
		return models.User{}, fmt.Errorf("error getting user by username: %w", result.Error)
	}
	return user, nil
}

// GetUserByID retrieves a user by their ID from the database using GORM
func (r *UserRepositoryImpl) GetUserByID(id uuid.UUID) (models.User, error) {
	var user models.User
	result := r.db.First(&user, id) // GORM can find by primary key directly
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.User{}, fmt.Errorf("user not found: %w", result.Error)
		}
		return models.User{}, fmt.Errorf("error getting user by ID: %w", result.Error)
	}
	return user, nil
}

// GetAllUsers retrieves all users from the database using GORM
func (r *UserRepositoryImpl) GetAllUsers() ([]models.User, error) {
	var users []models.User
	// Menggunakan Omit untuk tidak menyertakan password_hash dalam hasil
	result := r.db.Omit("password_hash").Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("error getting all users: %w", result.Error)
	}
	return users, nil
}

// CreateUser inserts a new user into the database using GORM
func (r *UserRepositoryImpl) CreateUser(user models.User) (models.User, error) {
	// GORM akan otomatis mengisi ID, CreatedAt, dan UpdatedAt
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	result := r.db.Create(&user)
	if result.Error != nil {
		return models.User{}, fmt.Errorf("failed to create user: %w", result.Error)
	}
	return user, nil
}

// UpdateUser updates an existing user in the database using GORM
func (r *UserRepositoryImpl) UpdateUser(user models.User) (models.User, error) {
	if user.ID == uuid.Nil {
		return models.User{}, errors.New("cannot update user without ID")
	}
	user.UpdatedAt = time.Now()

	// Gunakan Save untuk memperbarui semua field
	result := r.db.Save(&user)

	if result.Error != nil {
		return models.User{}, fmt.Errorf("failed to update user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return models.User{}, gorm.ErrRecordNotFound
	}
	return user, nil
}

// DeleteUser deletes a user from the database using GORM
func (r *UserRepositoryImpl) DeleteUser(id uuid.UUID) error {
	result := r.db.Delete(&models.User{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found for deletion")
	}
	return nil
}
