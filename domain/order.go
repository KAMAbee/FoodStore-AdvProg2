package domain

import (
    "time"
)

type Order struct {
    ID         string      `json:"id"`
    UserID     string      `json:"user_id"`
    Status     string      `json:"status"`
    TotalPrice float64     `json:"total_price"`
    CreatedAt  time.Time   `json:"created_at"`
    Items      []*OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
    ID        string   `json:"id"`
    OrderID   string   `json:"order_id"`
    ProductID string   `json:"product_id"`
    Quantity  int32    `json:"quantity"`
    Price     float64  `json:"price"`
    Product   *Product `json:"product,omitempty"`
}