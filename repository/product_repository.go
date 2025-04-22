package repository

import "AdvProg2/domain"

type ProductRepository interface {
    Create(product *domain.Product) error
    GetByID(id string) (*domain.Product, error)
    Update(product *domain.Product) error
    Delete(id string) error
    List(page, limit int32) ([]*domain.Product, int32, error)
    
    SearchByName(name string, page, limit int32) ([]*domain.Product, int32, error)
    SearchByPriceRange(minPrice, maxPrice float64, page, limit int32) ([]*domain.Product, int32, error)
    SearchByFilters(name string, minPrice, maxPrice float64, page, limit int32) ([]*domain.Product, int32, error)
}