package messaging

import (
    "AdvProg2/domain"
)

type MessageProducer interface {
    PublishOrderCreated(event domain.OrderCreatedEvent) error
    PublishProductCreated(event domain.ProductCreatedEvent) error
    Close() error
}