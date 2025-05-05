package domain

import "time"

type Message struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Data      []byte    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderCreatedEvent struct {
	OrderID    string           `json:"order_id"`
	UserID     string           `json:"user_id"`
	TotalPrice float64          `json:"total_price"`
	Items      []OrderItemEvent `json:"items"`
	CreatedAt  time.Time        `json:"created_at"`
}

type OrderItemEvent struct {
	ProductID string  `json:"product_id"`
	Quantity  int32   `json:"quantity"`
	Price     float64 `json:"price"`
}

type ProductCreatedEvent struct {
	ProductID string    `json:"product_id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Stock     int32     `json:"stock"`
	CreatedAt time.Time `json:"created_at"`
}

type ProductUpdatedEvent struct {
	ProductID string    `json:"product_id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Stock     int32     `json:"stock"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductDeletedEvent struct {
	ProductID string    `json:"product_id"`
	DeletedAt time.Time `json:"deleted_at"`
}
