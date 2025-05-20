package integration

import (
	"AdvProg2/domain"
	"AdvProg2/infrastructure/db"
	"AdvProg2/usecase"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestOrderCreationFlow(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	dbConn, err := db.NewPostgresConnection()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	orderRepo := db.NewPostgresOrderRepository(dbConn)
	productRepo := db.NewPostgresProductRepository(dbConn)

	orderUseCase := usecase.NewOrderUseCase(orderRepo, productRepo, nil)

	testProduct := &domain.Product{
		Name:  "Integration Test Product",
		Price: 25.99,
		Stock: 10,
	}

	err = productRepo.Create(testProduct)
	if err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	defer func() {
		productRepo.Delete(testProduct.ID)
	}()

	orderItems := []struct {
		ProductID string
		Quantity  int32
	}{
		{
			ProductID: testProduct.ID,
			Quantity:  2,
		},
	}

	testUserID := "integration-test-user"
	order, err := orderUseCase.CreateOrder(testUserID, orderItems)

	defer func() {
		if order != nil {
			orderRepo.Delete(order.ID)
		}
	}()

	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, testUserID, order.UserID)
	assert.Equal(t, "pending", order.Status)
	assert.Equal(t, 2*testProduct.Price, order.TotalPrice)
	assert.Equal(t, 1, len(order.Items))

	savedOrder, err := orderRepo.GetByID(order.ID)
	assert.NoError(t, err)
	assert.Equal(t, order.ID, savedOrder.ID)
	assert.Equal(t, order.UserID, savedOrder.UserID)
	assert.Equal(t, order.Status, savedOrder.Status)
	assert.Equal(t, order.TotalPrice, savedOrder.TotalPrice)

	updatedProduct, err := productRepo.GetByID(testProduct.ID)
	assert.NoError(t, err)
	assert.Equal(t, testProduct.Stock-2, updatedProduct.Stock)
}
