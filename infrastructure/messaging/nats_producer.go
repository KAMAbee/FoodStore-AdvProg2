package messaging

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

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

	log.Printf("Publishing order.created event for order %s", event.OrderID)
	err = p.nc.Publish(subject, msgBytes)
	if err != nil {
		log.Printf("Error publishing message: %v", err)
		return err
	}

	log.Printf("Published order.created event for order %s", event.OrderID)
	return nil
}

func (p *NatsProducer) PublishProductCreated(event domain.ProductCreatedEvent) error {
	subject := "product.created"

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshalling product created event: %v", err)
		return err
	}

	message := domain.Message{
		ID:        uuid.New().String(),
		Type:      "product.created",
		Data:      data,
		CreatedAt: time.Now(),
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return err
	}

	log.Printf("Publishing product.created event for product %s (%s)", event.ProductID, event.Name)
	err = p.nc.Publish(subject, msgBytes)
	if err != nil {
		log.Printf("Error publishing message: %v", err)
		return err
	}

	log.Printf("Successfully published product.created event for product %s", event.ProductID)
	return nil
}

func (p *NatsProducer) PublishProductUpdated(event domain.ProductUpdatedEvent) error {
	subject := "product.updated"

	log.Printf("NATS Producer: Marshalling product updated event for %s (%s)",
		event.ProductID, event.Name)

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshalling product updated event: %v", err)
		return err
	}

	message := domain.Message{
		ID:        uuid.New().String(),
		Type:      "product.updated",
		Data:      data,
		CreatedAt: time.Now(),
	}

	log.Printf("NATS Producer: Created message envelope with ID %s for product.updated",
		message.ID)

	msgBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return err
	}

	log.Printf("NATS Producer: Publishing product.updated event to subject %s", subject)
	err = p.nc.Publish(subject, msgBytes)
	if err != nil {
		log.Printf("Error publishing message: %v", err)
		return err
	}

	log.Printf("NATS Producer: Successfully published product.updated event for product %s (%s)",
		event.ProductID, event.Name)
	return nil
}

func (p *NatsProducer) PublishProductDeleted(event domain.ProductDeletedEvent) error {
	subject := "product.deleted"

	log.Printf("NATS Producer: Marshalling product deleted event for %s", event.ProductID)

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshalling product deleted event: %v", err)
		return err
	}

	message := domain.Message{
		ID:        uuid.New().String(),
		Type:      "product.deleted",
		Data:      data,
		CreatedAt: time.Now(),
	}

	log.Printf("NATS Producer: Created message envelope with ID %s for product.deleted",
		message.ID)

	msgBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return err
	}

	log.Printf("NATS Producer: Publishing product.deleted event to subject %s", subject)
	err = p.nc.Publish(subject, msgBytes)
	if err != nil {
		log.Printf("Error publishing message: %v", err)
		return err
	}

	log.Printf("NATS Producer: Successfully published product.deleted event for product %s",
		event.ProductID)
	return nil
}

func (p *NatsProducer) Close() error {
	p.nc.Close()
	return nil
}
