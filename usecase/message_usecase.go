package usecase

import (
	"errors"
	"log"
	"time"

	"AdvProg2/domain"
	"AdvProg2/pkg/cache"
	"AdvProg2/repository"
)

type MessageUseCase struct {
	producer    repository.MessageProducer
	productRepo repository.ProductRepository
	cache       *cache.Cache
}

func NewMessageUseCase(producer repository.MessageProducer, productRepo repository.ProductRepository, cache *cache.Cache) *MessageUseCase {
	return &MessageUseCase{
		producer:    producer,
		productRepo: productRepo,
		cache:       cache,
	}
}

func (uc *MessageUseCase) PublishOrderCreatedEvent(order *domain.Order) error {
	if uc.producer == nil {
		return errors.New("message producer not configured")
	}

	// Create order items 
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

func (uc *MessageUseCase) PublishProductCreatedEvent(product *domain.Product) error {
	if uc.producer == nil {
		return errors.New("message producer not configured")
	}

	event := domain.ProductCreatedEvent{
		ProductID: product.ID,
		Name:      product.Name,
		Price:     product.Price,
		Stock:     product.Stock,
		CreatedAt: time.Now(),
	}

	log.Printf("Message UseCase publishing product.created event: %+v", event)

	if err := uc.producer.PublishProductCreated(event); err != nil {
		log.Printf("Failed to publish product created event: %v", err)
		return err
	}

	return nil
}

func (uc *MessageUseCase) PublishProductUpdatedEvent(product *domain.Product) error {
	if uc.producer == nil {
		return errors.New("message producer not configured")
	}

	log.Printf("Creating product updated event for ID=%s, Name=%s, Price=%.2f, Stock=%d",
		product.ID, product.Name, product.Price, product.Stock)

	event := domain.ProductUpdatedEvent{
		ProductID: product.ID,
		Name:      product.Name,
		Price:     product.Price,
		Stock:     product.Stock,
		UpdatedAt: time.Now(),
	}

	log.Printf("Message UseCase publishing product.updated event: %+v", event)

	if err := uc.producer.PublishProductUpdated(event); err != nil {
		log.Printf("Failed to publish product updated event: %v", err)
		return err
	}

	log.Printf("Message UseCase successfully initiated product.updated event publishing")
	return nil
}

func (uc *MessageUseCase) PublishProductDeletedEvent(productID string) error {
	if uc.producer == nil {
		return errors.New("message producer not configured")
	}

	log.Printf("Creating product deleted event for ID=%s at %s",
		productID, time.Now().Format(time.RFC3339))

	event := domain.ProductDeletedEvent{
		ProductID: productID,
		DeletedAt: time.Now(),
	}

	log.Printf("Message UseCase publishing product.deleted event: %+v", event)

	if err := uc.producer.PublishProductDeleted(event); err != nil {
		log.Printf("Failed to publish product deleted event: %v", err)
		return err
	}

	log.Printf("Message UseCase successfully initiated product.deleted event publishing")
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

func (uc *MessageUseCase) HandleProductCreatedEvent(event domain.ProductCreatedEvent) error {
	log.Printf("Processing product created event for product %s", event.ProductID)
	log.Printf("New product added: %s, Price: $%.2f, Stock: %d", event.Name, event.Price, event.Stock)

	return nil
}

func (uc *MessageUseCase) HandleProductUpdatedEvent(event domain.ProductUpdatedEvent) error {
	log.Printf("Processing product updated event for product %s", event.ProductID)
	log.Printf("Product updated: %s, New price: $%.2f, New stock: %d, Updated at: %s",
		event.Name, event.Price, event.Stock, event.UpdatedAt.Format(time.RFC3339))

	product, err := uc.productRepo.GetByID(event.ProductID)
	if err != nil {
		log.Printf("Warning: Failed to find product %s in repository after update event: %v",
			event.ProductID, err)
	} else {
		log.Printf("Confirmed product %s exists in database after update", event.ProductID)

		// Update cache
		if uc.cache != nil {
			cacheKey := "product:" + event.ProductID
			uc.cache.Set(cacheKey, product, 5*time.Minute)
			log.Printf("Cache updated for product %s", event.ProductID)
		}
	}

	return nil
}

func (uc *MessageUseCase) HandleProductDeletedEvent(event domain.ProductDeletedEvent) error {
	log.Printf("Processing product deleted event for product %s", event.ProductID)
	log.Printf("Product %s was deleted at %s",
		event.ProductID, event.DeletedAt.Format(time.RFC3339))

	_, err := uc.productRepo.GetByID(event.ProductID)
	if err != nil {
		log.Printf("Confirmed product %s no longer exists in database after deletion",
			event.ProductID)
	} else {
		log.Printf("Warning: Product %s still exists in repository after delete event",
			event.ProductID)
	}

	if uc.cache != nil {
		cacheKey := "product:" + event.ProductID
		uc.cache.Delete(cacheKey)
		log.Printf("Cache invalidated for product %s", event.ProductID)
	}

	return nil
}
