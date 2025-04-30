package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcHandler "AdvProg2/handler/grpc"
	httpHandler "AdvProg2/handler/http"
	db "AdvProg2/infrastructure/db"
	"AdvProg2/infrastructure/messaging"
	pb "AdvProg2/proto/order"
	"AdvProg2/repository"
	"AdvProg2/usecase"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log.Println("Starting order service...")

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

	// Connect to NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	var messageProducer repository.MessageProducer
	var messageUseCase *usecase.MessageUseCase

	nc, err := messaging.NewNatsConnection(natsURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to NATS: %v", err)
		log.Println("Order service will run without messaging capabilities")
	} else {
		messageProducer = messaging.NewNatsProducer(nc)
		log.Println("Connected to NATS messaging system")
		defer nc.Close()
		defer messageProducer.Close()
	}

	orderRepo := db.NewPostgresOrderRepository(dbConn)
	productRepo := db.NewPostgresProductRepository(dbConn)

	// Create message use case only if we have a producer
	if messageProducer != nil {
		messageUseCase = usecase.NewMessageUseCase(messageProducer, productRepo)
		log.Println("Initialized message use case")
	}

	orderUseCase := usecase.NewOrderUseCase(orderRepo, productRepo, messageUseCase)
	log.Println("Initialized order use case")

	grpcOrderHandler := grpcHandler.NewOrderHandler(orderUseCase)

	grpcPort := os.Getenv("ORDER_SERVICE_PORT")
	if grpcPort == "" {
		grpcPort = "8083"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, grpcOrderHandler)

	reflection.Register(grpcServer)

	go func() {
		log.Printf("Order gRPC server started on port %s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	httpPort := os.Getenv("ORDER_SERVICE_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8093"
	}

	orderHTTPHandler := httpHandler.NewOrderHTTPHandler(orderUseCase)

	router := mux.NewRouter()

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	router.HandleFunc("/api/orders", orderHTTPHandler.CreateOrder).Methods("POST")
	router.HandleFunc("/api/orders/{id}", orderHTTPHandler.GetOrder).Methods("GET")
	router.HandleFunc("/api/orders", orderHTTPHandler.GetUserOrders).Methods("GET")
	router.HandleFunc("/api/orders/{id}", orderHTTPHandler.UpdateOrderStatus).Methods("PATCH")
	router.HandleFunc("/api/orders/{id}", orderHTTPHandler.CancelOrder).Methods("DELETE")

	router.HandleFunc("/api/orders", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("OPTIONS")
	router.HandleFunc("/api/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("OPTIONS")

	httpServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: router,
	}

	go func() {
		log.Printf("Order HTTP server started on port %s", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to serve HTTP: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server shutdown error: %v", err)
	}

	grpcServer.GracefulStop()
}
