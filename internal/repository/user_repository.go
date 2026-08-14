package repository

import (
	"errors"
	"fmt"

	"github.com/Rithvik-C18/go-flow/internal/database"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *database.User) error {
	if err := r.db.Create(user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("user %q: %w", user.Username, ErrConflict)
		}
		return err
	}
	return nil
}

func (r *UserRepository) FindByUsername(username string) (*database.User, error) {
	var user database.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user %q: %w", username, ErrNotFound)
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*database.User, error) {
	var user database.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user with email %q: %w", email, ErrNotFound)
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
