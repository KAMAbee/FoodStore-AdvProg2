package db

import (
	"AdvProg2/domain"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestPostgresOrderRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' ", err)
	}
	defer db.Close()

	repo := &PostgresOrderRepository{db: db}

	now := time.Now()
	order := &domain.Order{
		ID:         "test-order-id",
		UserID:     "test-user-id",
		Status:     "pending",
		TotalPrice: 100.0,
		CreatedAt:  now,
		Items: []*domain.OrderItem{
			{
				ID:        "test-item-id",
				OrderID:   "test-order-id",
				ProductID: "test-product-id",
				Quantity:  2,
				Price:     50.0,
			},
		},
	}

	mock.ExpectBegin()

	mock.ExpectExec("INSERT INTO orders").
		WithArgs(order.ID, order.UserID, order.Status, order.TotalPrice, order.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO order_items").
		WithArgs(order.Items[0].ID, order.Items[0].OrderID, order.Items[0].ProductID, order.Items[0].Quantity, order.Items[0].Price).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = repo.Create(order)

	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
