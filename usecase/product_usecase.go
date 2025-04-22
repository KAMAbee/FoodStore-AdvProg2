package usecase

import (
	"errors"
	"github.com/google/uuid"

	"AdvProg2/domain"
	"AdvProg2/repository"
)

type ProductUseCase struct {
	productRepo repository.ProductRepository
}

func (uc *ProductUseCase) SearchByPriceRange(minPrice float64, maxPrice float64, i int32, param4 int32) ([]*domain.Product, int32, error) {
	panic("unimplemented")
}

func NewProductUseCase(productRepo repository.ProductRepository) *ProductUseCase {
	return &ProductUseCase{
		productRepo: productRepo,
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
		return nil, errors.New("product price cannot be negative")
	}
	if stock < 0 {
		return nil, errors.New("product stock cannot be negative")
	}

	product := &domain.Product{
		ID:    uuid.New().String(),
		Name:  name,
		Price: price,
		Stock: stock,
	}

	if err := uc.productRepo.Create(product); err != nil {
		return nil, err
	}

	return product, nil
}

func (uc *ProductUseCase) UpdateProduct(id, name string, price float64, stock int32) (*domain.Product, error) {
	if id == "" {
		return nil, errors.New("product ID cannot be empty")
	}

	existingProduct, err := uc.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		existingProduct.Name = name
	}
	if price >= 0 {
		existingProduct.Price = price
	}
	if stock >= 0 {
		existingProduct.Stock = stock
	}

	if err := uc.productRepo.Update(existingProduct); err != nil {
		return nil, err
	}

	return existingProduct, nil
}

func (uc *ProductUseCase) DeleteProduct(id string) error {
	if id == "" {
		return errors.New("product ID cannot be empty")
	}

	_, err := uc.productRepo.GetByID(id)
	if err != nil {
		return err
	}

	return uc.productRepo.Delete(id)
}

func (uc *ProductUseCase) SearchProducts(name string, minPrice, maxPrice float64, page, limit int32) ([]*domain.Product, int32, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	if name != "" && (minPrice > 0 || maxPrice > 0) {
		return uc.productRepo.SearchByFilters(name, minPrice, maxPrice, page, limit)
	}

	if name != "" {
		return uc.productRepo.SearchByName(name, page, limit)
	}

	if minPrice > 0 || maxPrice > 0 {
		return uc.productRepo.SearchByPriceRange(minPrice, maxPrice, page, limit)
	}

	return uc.productRepo.List(page, limit)
}

