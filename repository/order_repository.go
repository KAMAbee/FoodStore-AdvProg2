package repository

import (
    "AdvProg2/domain"
)

type OrderRepository interface {
    Create(order *domain.Order) error
    GetByID(id string) (*domain.Order, error)
    GetByUserID(userID string, page, limit int32) ([]*domain.Order, int32, error)
    UpdateStatus(id, status string) error
    Delete(id string) error
}