package messaging

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"AdvProg2/domain"
)

type NatsConsumer struct {
	nc            *nats.Conn
	subscriptions []*nats.Subscription
}

func NewNatsConsumer(nc *nats.Conn) *NatsConsumer {
	return &NatsConsumer{
		nc:            nc,
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

func (c *NatsConsumer) SubscribeToProductCreated(handler func(event domain.ProductCreatedEvent) error) error {
	subject := "product.created"

	log.Printf("Subscribing to %s", subject)

	subscription, err := c.nc.Subscribe(subject, func(m *nats.Msg) {

		var message domain.Message
		if err := json.Unmarshal(m.Data, &message); err != nil {
			return
		}

		var event domain.ProductCreatedEvent
		if err := json.Unmarshal(message.Data, &event); err != nil {
			log.Printf("Event data content: %s", string(message.Data))
			return
		}

		log.Printf("Successfully parsed product.created event for product %s (%s)",
			event.ProductID, event.Name)

		if err := handler(event); err != nil {
			log.Printf("Error handling product created event: %v", err)
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

func (c *NatsConsumer) SubscribeToProductUpdated(handler func(event domain.ProductUpdatedEvent) error) error {
	subject := "product.updated"

	log.Printf("NATS Consumer: Subscribing to %s", subject)

	subscription, err := c.nc.Subscribe(subject, func(m *nats.Msg) {
		log.Printf("NATS Consumer: Received raw message on %s: %s", subject, string(m.Data))

		var message domain.Message
		if err := json.Unmarshal(m.Data, &message); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			log.Printf("Raw message content: %s", string(m.Data))
			return
		}

		log.Printf("NATS Consumer: Message envelope parsed: ID=%s, Type=%s",
			message.ID, message.Type)

		// Use message.Data directly as it's already []byte
		eventData := message.Data
		log.Printf("NATS Consumer: Using message data as bytes")

		var event domain.ProductUpdatedEvent
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Printf("Error unmarshalling product updated event: %v", err)
			log.Printf("Event data content: %s", string(eventData))
			return
		}

		log.Printf("NATS Consumer: Successfully parsed product.updated event for product %s (%s)",
			event.ProductID, event.Name)
		log.Printf("NATS Consumer: Product details: Price=$%.2f, Stock=%d, Updated at %s",
			event.Price, event.Stock, event.UpdatedAt.Format(time.RFC3339))

		if err := handler(event); err != nil {
			log.Printf("Error handling product updated event: %v", err)
		}
	})

	if err != nil {
		log.Printf("Error subscribing to %s: %v", subject, err)
		return err
	}

	c.subscriptions = append(c.subscriptions, subscription)

	log.Printf("NATS Consumer: Successfully subscribed to %s", subject)
	return nil
}

func (c *NatsConsumer) SubscribeToProductDeleted(handler func(event domain.ProductDeletedEvent) error) error {
	subject := "product.deleted"

	log.Printf("NATS Consumer: Subscribing to %s", subject)

	subscription, err := c.nc.Subscribe(subject, func(m *nats.Msg) {
		log.Printf("NATS Consumer: Received raw message on %s: %s", subject, string(m.Data))

		var message domain.Message
		if err := json.Unmarshal(m.Data, &message); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			log.Printf("Raw message content: %s", string(m.Data))
			return
		}

		log.Printf("NATS Consumer: Message envelope parsed: ID=%s, Type=%s",
			message.ID, message.Type)

		// Use message.Data directly as it's already []byte
		eventData := message.Data
		log.Printf("NATS Consumer: Using message data as bytes")

		var event domain.ProductDeletedEvent
		if err := json.Unmarshal(eventData, &event); err != nil {
			log.Printf("Error unmarshalling product deleted event: %v", err)
			log.Printf("Event data content: %s", string(eventData))
			return
		}

		log.Printf("NATS Consumer: Successfully parsed product.deleted event for product %s",
			event.ProductID)
		log.Printf("NATS Consumer: Product was deleted at %s",
			event.DeletedAt.Format(time.RFC3339))

		if err := handler(event); err != nil {
			log.Printf("Error handling product deleted event: %v", err)
		}
	})

	if err != nil {
		log.Printf("Error subscribing to %s: %v", subject, err)
		return err
	}

	c.subscriptions = append(c.subscriptions, subscription)

	log.Printf("NATS Consumer: Successfully subscribed to %s", subject)
	return nil
}

func (c *NatsConsumer) Close() error {
	for _, sub := range c.subscriptions {
		sub.Unsubscribe()
	}
	c.nc.Close()
	return nil
}
