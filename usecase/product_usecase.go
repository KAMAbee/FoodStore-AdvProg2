package usecase

import (
	"errors"

	"github.com/google/uuid"

	"AdvProg2/domain"
	"AdvProg2/repository"
)

type ProductUseCase struct {
	productRepo    repository.ProductRepository
	messageUseCase *MessageUseCase
}

func NewProductUseCase(productRepo repository.ProductRepository, messageUseCase *MessageUseCase) *ProductUseCase {
	return &ProductUseCase{
		productRepo:    productRepo,
		messageUseCase: messageUseCase,
	}
}

func (uc *ProductUseCase) GetProduct(id string) (*domain.Product, error) {
	if id == "" {
		return nil, errors.New("product ID cannot be empty")
	}

	return uc.productRepo.GetByID(id)
}

func (uc *ProductUseCase) ListProducts(page, limit int32) ([]*domain.Product, int32, error) {
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 10
	}

	return uc.productRepo.List(page, limit)
}

func (uc *ProductUseCase) CreateProduct(name string, price float64, stock int32) (*domain.Product, error) {
	if name == "" {
		return nil, errors.New("product name cannot be empty")
	}

	if price < 0 {
		return nil, errors.New("price cannot be negative")
	}

	if stock < 0 {
		return nil, errors.New("stock cannot be negative")
	}

	product := &domain.Product{
		ID:        uuid.New().String(),
		Name:      name,
		Price:     price,
		Stock:     stock,
	}

	err := uc.productRepo.Create(product)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (uc *ProductUseCase) UpdateProduct(id, name string, price float64, stock int32) (*domain.Product, error) {
	if id == "" {
		return nil, errors.New("product ID cannot be empty")
	}

	product, err := uc.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		product.Name = name
	}

	if price >= 0 {
		product.Price = price
	}

	if stock >= 0 {
		product.Stock = stock
	}


	err = uc.productRepo.Update(product)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (uc *ProductUseCase) DeleteProduct(id string) error {
	if id == "" {
		return errors.New("product ID cannot be empty")
	}

	// Check if product exists
	_, err := uc.productRepo.GetByID(id)
	if err != nil {
		return err
	}

	return uc.productRepo.Delete(id)
}

func (uc *ProductUseCase) SearchByPriceRange(minPrice float64, maxPrice float64, page int32, limit int32) ([]*domain.Product, int32, error) {
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 10
	}

	filters := make(map[string]interface{})

	if minPrice > 0 {
		filters["min_price"] = minPrice
	}

	if maxPrice > 0 {
		filters["max_price"] = maxPrice
	}

	return uc.productRepo.List(page, limit)
}
