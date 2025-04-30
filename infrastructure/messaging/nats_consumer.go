package messaging

import (
    "encoding/json"
    "github.com/nats-io/nats.go"
    "log"

    "AdvProg2/domain"
)

type NatsConsumer struct {
    nc           *nats.Conn
    subscriptions []*nats.Subscription
}

func NewNatsConsumer(nc *nats.Conn) *NatsConsumer {
    return &NatsConsumer{
        nc:           nc,
        subscriptions: make([]*nats.Subscription, 0),
    }
}

func (c *NatsConsumer) SubscribeToOrderCreated(handler func(event domain.OrderCreatedEvent) error) error {
    subject := "order.created"
    
    log.Printf("Subscribing to %s", subject)
    
    subscription, err := c.nc.Subscribe(subject, func(m *nats.Msg) {
        var message domain.Message
        if err := json.Unmarshal(m.Data, &message); err != nil {
            log.Printf("Error unmarshalling message: %v", err)
            return
        }

        var event domain.OrderCreatedEvent
        if err := json.Unmarshal(message.Data, &event); err != nil {
            log.Printf("Error unmarshalling order created event: %v", err)
            return
        }

        log.Printf("Received order.created event for order %s", event.OrderID)
        
        if err := handler(event); err != nil {
            log.Printf("Error handling order created event: %v", err)
        }
    })

    if err != nil {
        log.Printf("Error subscribing to %s: %v", subject, err)
        return err
    }

    c.subscriptions = append(c.subscriptions, subscription)
    
    log.Printf("Successfully subscribed to %s", subject)
    return nil
}

func (c *NatsConsumer) Close() error {
    for _, sub := range c.subscriptions {
        sub.Unsubscribe()
    }
    c.nc.Close()
    return nil
}