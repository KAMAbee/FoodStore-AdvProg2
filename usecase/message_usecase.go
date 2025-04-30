package usecase

import (
    "errors"
    "log"

    "AdvProg2/domain"
    "AdvProg2/repository"
)

type MessageUseCase struct {
    producer    repository.MessageProducer
    productRepo repository.ProductRepository
}

func NewMessageUseCase(producer repository.MessageProducer, productRepo repository.ProductRepository) *MessageUseCase {
    return &MessageUseCase{
        producer:    producer,
        productRepo: productRepo,
    }
}

func (uc *MessageUseCase) PublishOrderCreatedEvent(order *domain.Order) error {
    if uc.producer == nil {
        return errors.New("message producer not configured")
    }

    itemEvents := make([]domain.OrderItemEvent, 0, len(order.Items))
    for _, item := range order.Items {
        itemEvents = append(itemEvents, domain.OrderItemEvent{
            ProductID: item.ProductID,
            Quantity:  item.Quantity,
            Price:     item.Price,
        })
    }

    event := domain.OrderCreatedEvent{
        OrderID:    order.ID,
        UserID:     order.UserID,
        TotalPrice: order.TotalPrice,
        Items:      itemEvents,
        CreatedAt:  order.CreatedAt,
    }

    if err := uc.producer.PublishOrderCreated(event); err != nil {
        log.Printf("Failed to publish order created event: %v", err)
        return err
    }

    return nil
}

func (uc *MessageUseCase) HandleOrderCreatedEvent(event domain.OrderCreatedEvent) error {
    log.Printf("Processing order created event for order %s", event.OrderID)
    log.Printf("User %s created an order for $%.2f", event.UserID, event.TotalPrice)
    
    for _, item := range event.Items {
        product, err := uc.productRepo.GetByID(item.ProductID)
        if err != nil {
            log.Printf("Error getting product %s: %v", item.ProductID, err)
            continue
        }
        
        log.Printf("Decreasing stock for product %s (%s) by %d units", 
            product.ID, product.Name, item.Quantity)
        
    }
    
    return nil
}