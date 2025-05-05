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

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	grpcHandler "AdvProg2/handler/grpc"
	httpHandler "AdvProg2/handler/http"
	"AdvProg2/infrastructure/db"
	"AdvProg2/infrastructure/messaging"
	pb "AdvProg2/proto/user"
	"AdvProg2/repository"
	"AdvProg2/usecase"
)

func main() {
	log.Println("Starting user service...")

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
	userRepo, err := db.NewPostgresUserRepository(dbConn)
	if err != nil {
		log.Fatalf("Failed to create user repository: %v", err)
	}

	productRepo := db.NewPostgresProductRepository(dbConn)
	log.Println("Initialized repositories")

	// Connect to NATS for admin-product messaging
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	var messageProducer repository.MessageProducer
	var messageUseCase *usecase.MessageUseCase

	nc, err := messaging.NewNatsConnection(natsURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to NATS: %v", err)
		log.Println("User service will run without messaging capabilities")
	} else {
		messageProducer = messaging.NewNatsProducer(nc)
		messageUseCase = usecase.NewMessageUseCase(messageProducer, productRepo)
		log.Println("Connected to NATS messaging system")
		defer nc.Close()
		defer messageProducer.Close()
	}

	// Create use cases
	userUseCase := usecase.NewUserUseCase(userRepo)
	productUseCase := usecase.NewProductUseCase(productRepo, messageUseCase)
	log.Println("Initialized use cases")

	// Setup gRPC handler
	grpcUserHandler := grpcHandler.NewUserHandler(userUseCase)

	grpcPort := os.Getenv("USER_SERVICE_PORT")
	if grpcPort == "" {
		grpcPort = "8084"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, grpcUserHandler)

	reflection.Register(grpcServer)

	go func() {
		log.Printf("User gRPC server started on port %s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Setup HTTP server
	httpPort := os.Getenv("USER_SERVICE_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8085"
	}

	userHTTPHandler := httpHandler.NewUserHTTPHandler(userUseCase)

	// Create admin handler
	adminHTTPHandler := httpHandler.NewAdminHTTPHandler(productUseCase, messageUseCase)
	log.Println("Initialized admin HTTP handler")

	router := mux.NewRouter()

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-Role")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// User service endpoints
	router.HandleFunc("/api/users/register", userHTTPHandler.Register).Methods("POST")
	router.HandleFunc("/api/users/login", userHTTPHandler.Login).Methods("POST")
	router.HandleFunc("/api/users/profile/{id}", userHTTPHandler.GetProfile).Methods("GET")

	// Set up admin routes
	router.HandleFunc("/api/admin/products", adminHTTPHandler.CreateProduct).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/admin/products/{id}", adminHTTPHandler.UpdateProduct).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/admin/products/{id}", adminHTTPHandler.DeleteProduct).Methods("DELETE", "OPTIONS")

	httpServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: router,
	}

	go func() {
		log.Printf("User HTTP server started on port %s", httpPort)
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

	log.Println("User service gracefully stopped")
}
