package usecase

import (
    "errors"

    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
    
    "AdvProg2/domain"
    "AdvProg2/repository"
)

const (
    bcryptCost = 12
    minPasswordLength = 6
)

type UserUseCase struct {
    userRepo repository.UserRepository
}

func NewUserUseCase(userRepo repository.UserRepository) *UserUseCase {
    return &UserUseCase{
        userRepo: userRepo,
    }
}

func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
    return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

func (uc *UserUseCase) Register(username, password string) (*domain.User, error) {
    if username == "" {
        return nil, errors.New("username is required")
    }
    if len(password) < minPasswordLength {
        return nil, errors.New("password must be at least 6 characters long")
    }

    _, err := uc.userRepo.GetByUsername(username)
    if err == nil {
        return nil, repository.ErrUsernameAlreadyExists
    } else if err != repository.ErrUserNotFound {
        return nil, err
    }

    hashedPassword, err := hashPassword(password)
    if err != nil {
        return nil, err
    }

    user := &domain.User{
        ID:       uuid.New().String(),
        Username: username,
        Password: hashedPassword,
    }

    err = uc.userRepo.Create(user)
    if err != nil {
        return nil, err
    }

    userResponse := *user
    userResponse.Password = ""
    return &userResponse, nil
}

func (uc *UserUseCase) Login(username, password string) (*domain.User, error) {
    if username == "" || password == "" {
        return nil, repository.ErrInvalidCredentials
    }

    user, err := uc.userRepo.GetByUsername(username)
    if err != nil {
        if err == repository.ErrUserNotFound {
            return nil, repository.ErrInvalidCredentials
        }
        return nil, err
    }

    if !checkPasswordHash(password, user.Password) {
        return nil, repository.ErrInvalidCredentials
    }

    userResponse := *user
    userResponse.Password = ""
    return &userResponse, nil
}

func (uc *UserUseCase) GetProfile(id string) (*domain.User, error) {
    if id == "" {
        return nil, errors.New("user ID is required")
    }

    user, err := uc.userRepo.GetByID(id)
    if err != nil {
        return nil, err
    }

    userResponse := *user
    userResponse.Password = ""
    return &userResponse, nil
}