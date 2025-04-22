package repository

import (
	"errors"
    "AdvProg2/domain"
)
var (
    ErrUserNotFound         = errors.New("user not found")
    ErrUsernameAlreadyExists = errors.New("username already exists")
    ErrInvalidCredentials   = errors.New("invalid credentials")
)

type UserRepository interface {
    Create(user *domain.User) error 
    GetByID(id string) (*domain.User, error)
    GetByUsername(username string) (*domain.User, error)
}