package usecase

import (
    "errors"
    "time"
    
    "github.com/google/uuid"
    "AdvProg2/domain"
    "AdvProg2/repository"
)

type OrderUseCase struct {
    orderRepo   repository.OrderRepository
    productRepo repository.ProductRepository
}

func NewOrderUseCase(orderRepo repository.OrderRepository, productRepo repository.ProductRepository) *OrderUseCase {
    return &OrderUseCase{
        orderRepo:   orderRepo,
        productRepo: productRepo,
    }
}

func (uc *OrderUseCase) CreateOrder(userID string, orderItems []struct{ProductID string; Quantity int32}) (*domain.Order, error) {
    if userID == "" {
        return nil, errors.New("user ID cannot be empty")
    }
    
    if len(orderItems) == 0 {
        return nil, errors.New("order must have at least one item")
    }
    
    var totalPrice float64
    var orderItemsEntities []*domain.OrderItem
    
    for _, item := range orderItems {
        if item.Quantity <= 0 {
            return nil, errors.New("product quantity must be positive")
        }
        
        product, err := uc.productRepo.GetByID(item.ProductID)
        if err != nil {
            return nil, err
        }
        
        if product.Stock < item.Quantity {
            return nil, errors.New("not enough stock for product: " + product.Name)
        }
        
        product.Stock -= item.Quantity
        err = uc.productRepo.Update(product)
        if err != nil {
            return nil, err
        }
        
        orderItem := &domain.OrderItem{
            ID:        uuid.New().String(),
            ProductID: product.ID,
            Quantity:  item.Quantity,
            Price:     product.Price,
            Product:   product,
        }
        
        orderItemsEntities = append(orderItemsEntities, orderItem)
        totalPrice += product.Price * float64(item.Quantity)
    }
    
    order := &domain.Order{
        ID:         uuid.New().String(),
        UserID:     userID,
        Status:     "pending",
        TotalPrice: totalPrice,
        CreatedAt:  time.Now(),
        Items:      orderItemsEntities,
    }
    
    err := uc.orderRepo.Create(order)
    if err != nil {
        return nil, err
    }
    
    return order, nil
}

func (uc *OrderUseCase) GetOrder(id string) (*domain.Order, error) {
    if id == "" {
        return nil, errors.New("order ID cannot be empty")
    }
    
    return uc.orderRepo.GetByID(id)
}

func (uc *OrderUseCase) GetUserOrders(userID string, page, limit int32) ([]*domain.Order, int32, error) {
    if userID == "" {
        return nil, 0, errors.New("user ID cannot be empty")
    }
    
    if page <= 0 {
        page = 1
    }
    
    if limit <= 0 {
        limit = 10
    }
    
    return uc.orderRepo.GetByUserID(userID, page, limit)
}

func (uc *OrderUseCase) UpdateOrderStatus(id, status string) (*domain.Order, error) {
    if id == "" {
        return nil, errors.New("order ID cannot be empty")
    }
    
    if status == "" {
        return nil, errors.New("status cannot be empty")
    }
    
    validStatuses := map[string]bool{
        "pending":   true,
        "completed": true,
        "cancelled": true,
    }
    
    if !validStatuses[status] {
        return nil, errors.New("invalid status")
    }
    
    order, err := uc.orderRepo.GetByID(id)
    if err != nil {
        return nil, err
    }
    
    if order.Status == "completed" || order.Status == "cancelled" {
        return nil, errors.New("cannot change status of a completed or cancelled order")
    }
    
    if status == "cancelled" && order.Status != "cancelled" {
        for _, item := range order.Items {
            product, err := uc.productRepo.GetByID(item.ProductID)
            if err != nil {
                return nil, err
            }
            
            product.Stock += item.Quantity
            err = uc.productRepo.Update(product)
            if err != nil {
                return nil, err
            }
        }
    }
    
    err = uc.orderRepo.UpdateStatus(id, status)
    if err != nil {
        return nil, err
    }
    
    return uc.orderRepo.GetByID(id)
}

func (uc *OrderUseCase) CancelOrder(id string) error {
    if id == "" {
        return errors.New("order ID cannot be empty")
    }
    
    _, err := uc.UpdateOrderStatus(id, "cancelled")
    return err
}