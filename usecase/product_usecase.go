package usecase

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"AdvProg2/domain"
	"AdvProg2/pkg/cache"
	"AdvProg2/repository"
)

type ProductUseCase struct {
	productRepo    repository.ProductRepository
	messageUseCase *MessageUseCase
	cache          *cache.Cache
}

func NewProductUseCase(productRepo repository.ProductRepository, messageUseCase *MessageUseCase) *ProductUseCase {
	return &ProductUseCase{
		productRepo:    productRepo,
		messageUseCase: messageUseCase,
		cache:          cache.New(),
	}
}

func (uc *ProductUseCase) GetProduct(id string) (*domain.Product, error) {
	if id == "" {
		return nil, errors.New("product ID cannot be empty")
	}

	cacheKey := "product:" + id
	if cachedData, found := uc.cache.Get(cacheKey); found {
		if product, ok := cachedData.(*domain.Product); ok {
			return product, nil
		}

		if jsonMap, ok := cachedData.(map[string]interface{}); ok {
			product := &domain.Product{ID: id}

			if name, ok := jsonMap["Name"]; ok && name != nil {
				if nameStr, ok := name.(string); ok {
					product.Name = nameStr
				}
			}

			if price, ok := jsonMap["Price"]; ok && price != nil {
				if priceFloat, ok := price.(float64); ok {
					product.Price = priceFloat
				}
			}

			if stock, ok := jsonMap["Stock"]; ok && stock != nil {
				if stockFloat, ok := stock.(float64); ok {
					product.Stock = int32(stockFloat)
				}
			}

			return product, nil
		}

		jsonData, err := json.Marshal(cachedData)
		if err == nil {
			var product domain.Product
			if err = json.Unmarshal(jsonData, &product); err == nil {
				return &product, nil
			}
		}

		uc.cache.Delete(cacheKey)
	}

	product, err := uc.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	uc.cache.Set(cacheKey, product, 5*time.Minute)

	return product, nil
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
		ID:    uuid.New().String(),
		Name:  name,
		Price: price,
		Stock: stock,
	}

	err := uc.productRepo.Create(product)
	if err != nil {
		return nil, err
	}

	// Add to cache
	cacheKey := "product:" + product.ID
	uc.cache.Set(cacheKey, product, 5*time.Minute)

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

	cacheKey := "product:" + id
	uc.cache.Set(cacheKey, product, 5*time.Minute)

	return product, nil
}

func (uc *ProductUseCase) DeleteProduct(id string) error {
	if id == "" {
		return errors.New("product ID cannot be empty")
	}

	_, err := uc.productRepo.GetByID(id)
	if err != nil {
		return err
	}

	err = uc.productRepo.Delete(id)
	if err != nil {
		return err
	}

	// Remove from cache
	cacheKey := "product:" + id
	uc.cache.Delete(cacheKey)

	return nil
}

func (uc *ProductUseCase) SearchByName(name string, page, limit int32) ([]*domain.Product, int32, error) {
	return uc.productRepo.SearchByName(name, page, limit)
}

func (uc *ProductUseCase) SearchByPriceRange(minPrice, maxPrice float64, page, limit int32) ([]*domain.Product, int32, error) {
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

	return uc.productRepo.SearchByPriceRange(minPrice, maxPrice, page, limit)
}
