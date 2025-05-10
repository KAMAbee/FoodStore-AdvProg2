package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/joho/godotenv"
    "github.com/nats-io/nats.go"

    "AdvProg2/domain"
    "AdvProg2/pkg/cache"
    "AdvProg2/infrastructure/db"
    "AdvProg2/infrastructure/messaging"
    "AdvProg2/usecase"
)

func main() {
    log.Println("Starting consumer service...")
    
    err := godotenv.Load()
    if err != nil {
        log.Printf("Warning: Error loading .env file: %v", err)
    }

    dbConn, err := db.NewPostgresConnection()
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer dbConn.Close()
    log.Println("Connected to database")

    productRepo := db.NewPostgresProductRepository(dbConn)
    log.Println("Initialized product repository")

    natsURL := os.Getenv("NATS_URL")
    if natsURL == "" {
        natsURL = nats.DefaultURL
    }

    nc, err := messaging.NewNatsConnection(natsURL)
    if err != nil {
        log.Fatalf("Failed to connect to NATS: %v", err)
    }
    defer nc.Close()
    consumer := messaging.NewNatsConsumer(nc)
    defer consumer.Close()
    
    producer := messaging.NewNatsProducer(nc)
    
    // Initialize cache
    cacheInstance := cache.New()

    messageUseCase := usecase.NewMessageUseCase(producer, productRepo, cacheInstance)

    err = consumer.SubscribeToOrderCreated(func(event domain.OrderCreatedEvent) error {
        return messageUseCase.HandleOrderCreatedEvent(event)
    })

    if err != nil {
        log.Fatalf("Failed to subscribe to order.created events: %v", err)
    }

    log.Println("Consumer service started, listening for order.created events")

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Consumer service shutting down")
}