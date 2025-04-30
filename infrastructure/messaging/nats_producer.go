package messaging

import (
    "encoding/json"
    "github.com/google/uuid"
    "github.com/nats-io/nats.go"
    "log"
    "time"

    "AdvProg2/domain"
)

type NatsProducer struct {
    nc *nats.Conn
}

func NewNatsProducer(nc *nats.Conn) *NatsProducer {
    return &NatsProducer{
        nc: nc,
    }
}

func (p *NatsProducer) PublishOrderCreated(event domain.OrderCreatedEvent) error {
    subject := "order.created"
    
    data, err := json.Marshal(event)
    if err != nil {
        log.Printf("Error marshalling order created event: %v", err)
        return err
    }

    message := domain.Message{
        ID:        uuid.New().String(),
        Type:      "order.created",
        Data:      data,
        CreatedAt: time.Now(),
    }

    msgBytes, err := json.Marshal(message)
    if err != nil {
        log.Printf("Error marshalling message: %v", err)
        return err
    }

    err = p.nc.Publish(subject, msgBytes)
    if err != nil {
        log.Printf("Error publishing message: %v", err)
        return err
    }

    log.Printf("Published order.created event for order %s", event.OrderID)
    return nil
}

func (p *NatsProducer) Close() error {
    p.nc.Close()
    return nil
}