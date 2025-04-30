package repository

import "AdvProg2/domain"

type MessageProducer interface {
    PublishOrderCreated(event domain.OrderCreatedEvent) error
    Close() error
}

type MessageConsumer interface {
    SubscribeToOrderCreated(handler func(event domain.OrderCreatedEvent) error) error
    Close() error
}