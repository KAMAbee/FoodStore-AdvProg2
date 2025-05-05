package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"

	"AdvProg2/domain"
	"AdvProg2/infrastructure/db"
	"AdvProg2/infrastructure/messaging"
	"AdvProg2/usecase"
)

func main() {
	log.Println("Starting admin consumer service...")

	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Connect to database
	dbConn, err := db.NewPostgresConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()
	log.Println("Connected to database")

	// Create repositories
	productRepo := db.NewPostgresProductRepository(dbConn)
	log.Println("Initialized product repository")

	// Connect to NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	nc, err := messaging.NewNatsConnection(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()
	log.Println("Connected to NATS messaging system")

	consumer := messaging.NewNatsConsumer(nc)
	defer consumer.Close()

	messageUseCase := usecase.NewMessageUseCase(nil, productRepo)

	err = consumer.SubscribeToProductCreated(func(event domain.ProductCreatedEvent) error {
		log.Printf("Admin consumer: received product created event for product %s", event.ProductID)
		return messageUseCase.HandleProductCreatedEvent(event)
	})
	if err != nil {
		log.Printf("Failed to subscribe to product.created events: %v", err)
	}

	err = consumer.SubscribeToProductUpdated(func(event domain.ProductUpdatedEvent) error {
		log.Printf("Admin consumer: received product.updated event for product %s (%s)",
			event.ProductID, event.Name)
		log.Printf("Admin consumer: product %s updated with price $%.2f and stock %d at %s",
			event.ProductID, event.Price, event.Stock, event.UpdatedAt.Format(time.RFC3339))
		return messageUseCase.HandleProductUpdatedEvent(event)
	})
	if err != nil {
		log.Printf("Failed to subscribe to product.updated events: %v", err)
	}

	err = consumer.SubscribeToProductDeleted(func(event domain.ProductDeletedEvent) error {
		log.Printf("Admin consumer: received product.deleted event for product %s", event.ProductID)
		log.Printf("Admin consumer: product %s was deleted at %s",
			event.ProductID, event.DeletedAt.Format(time.RFC3339))
		return messageUseCase.HandleProductDeletedEvent(event)
	})
	if err != nil {
		log.Printf("Failed to subscribe to product.deleted events: %v", err)
	}

	log.Println("Admin consumer started, listening for product events...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Admin consumer service shutting down")
}
