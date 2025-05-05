package repository

import "AdvProg2/domain"

type MessageProducer interface {
	PublishOrderCreated(event domain.OrderCreatedEvent) error
	PublishProductCreated(event domain.ProductCreatedEvent) error
	PublishProductUpdated(event domain.ProductUpdatedEvent) error
	PublishProductDeleted(event domain.ProductDeletedEvent) error
	Close() error
}

type MessageConsumer interface {
	SubscribeToOrderCreated(handler func(event domain.OrderCreatedEvent) error) error
	SubscribeToProductCreated(handler func(event domain.ProductCreatedEvent) error) error
	SubscribeToProductUpdated(handler func(event domain.ProductUpdatedEvent) error) error
	SubscribeToProductDeleted(handler func(event domain.ProductDeletedEvent) error) error
	Close() error
}
