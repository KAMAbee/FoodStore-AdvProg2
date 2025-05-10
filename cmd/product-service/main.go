package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"AdvProg2/domain"
	grpcHandler "AdvProg2/handler/grpc"
	"AdvProg2/pkg/cache"
	"AdvProg2/infrastructure/db"
	"AdvProg2/infrastructure/messaging"
	pb "AdvProg2/proto/product"
	"AdvProg2/usecase"
)

type ProductHTTPHandler struct {
	productUseCase *usecase.ProductUseCase
}

func NewProductHTTPHandler(productUseCase *usecase.ProductUseCase) *ProductHTTPHandler {
	return &ProductHTTPHandler{
		productUseCase: productUseCase,
	}
}

func (h *ProductHTTPHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	page := 1
	limit := 10
	var minPrice, maxPrice float64

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("per_page"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if minPriceStr := r.URL.Query().Get("min_price"); minPriceStr != "" {
		if mp, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			minPrice = mp
		}
	}

	if maxPriceStr := r.URL.Query().Get("max_price"); maxPriceStr != "" {
		if mp, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			maxPrice = mp
		}
	}

	var products []*domain.Product
	var total int32
	var err error

	if minPrice > 0 || maxPrice > 0 {
		products, total, err = h.productUseCase.SearchByPriceRange(minPrice, maxPrice, int32(page), int32(limit))
	} else {
		products, total, err = h.productUseCase.ListProducts(int32(page), int32(limit))
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"products": products,
		"total":    total,
		"page":     page,
		"per_page": limit,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *ProductHTTPHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	product, err := h.productUseCase.GetProduct(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(product)
}

func (h *ProductHTTPHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var product domain.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	createdProduct, err := h.productUseCase.CreateProduct(product.Name, product.Price, product.Stock)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdProduct)
}

func (h *ProductHTTPHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	var product domain.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updatedProduct, err := h.productUseCase.UpdateProduct(id, product.Name, product.Price, product.Stock)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updatedProduct)
}

func (h *ProductHTTPHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.productUseCase.DeleteProduct(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	log.Println("Starting product service...")

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

	// Add this near your other initializations:
	productCache := cache.New()

	// Connect to NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	var messageUseCase *usecase.MessageUseCase

	nc, err := messaging.NewNatsConnection(natsURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to NATS: %v", err)
		log.Println("Product service will run without messaging capabilities")
	} else {
		consumer := messaging.NewNatsConsumer(nc)
		messageUseCase = usecase.NewMessageUseCase(nil, productRepo, productCache)

		log.Println("Subscribing to product.created")
		err = consumer.SubscribeToProductCreated(func(event domain.ProductCreatedEvent) error {
			log.Printf("Received product.created event for product %s", event.ProductID)
			log.Printf("Processing product created event for product %s", event.ProductID)
			log.Printf("Admin user created product %s (%s) with price $%.2f and stock %d",
				event.ProductID, event.Name, event.Price, event.Stock)
			return messageUseCase.HandleProductCreatedEvent(event)
		})
		if err != nil {
			log.Printf("Warning: Failed to subscribe to product.created events: %v", err)
		} else {
			log.Println("Successfully subscribed to product.created")
		}

		log.Println("Subscribing to product.updated")
		err = consumer.SubscribeToProductUpdated(func(event domain.ProductUpdatedEvent) error {
			log.Printf("Received product.updated event for product %s", event.ProductID)
			log.Printf("Processing product updated event for product %s", event.ProductID)
			log.Printf("Admin user updated product %s (%s) to price $%.2f and stock %d",
				event.ProductID, event.Name, event.Price, event.Stock)

			result := messageUseCase.HandleProductUpdatedEvent(event)
			if result == nil {
				log.Printf("Successfully processed product.updated event for %s", event.ProductID)
			}
			return result
		})
		if err != nil {
			log.Printf("Warning: Failed to subscribe to product.updated events: %v", err)
		} else {
			log.Println("Successfully subscribed to product.updated")
		}

		log.Println("Subscribing to product.deleted")
		err = consumer.SubscribeToProductDeleted(func(event domain.ProductDeletedEvent) error {
			log.Printf("Received product.deleted event for product %s", event.ProductID)
			log.Printf("Processing product deleted event for product %s", event.ProductID)
			log.Printf("Admin user deleted product %s at %s",
				event.ProductID, event.DeletedAt.Format(time.RFC3339))

			result := messageUseCase.HandleProductDeletedEvent(event)
			if result == nil {
				log.Printf("Successfully processed product.deleted event for %s", event.ProductID)
			}
			return result
		})
		if err != nil {
			log.Printf("Warning: Failed to subscribe to product.deleted events: %v", err)
		} else {
			log.Println("Successfully subscribed to product.deleted")
		}

		log.Println("Product service started, listening for product events")
		defer nc.Close()
		defer consumer.Close()
	}

	productUseCase := usecase.NewProductUseCase(productRepo, messageUseCase)

	grpcProductHandler := grpcHandler.NewProductHandler(productUseCase)

	grpcPort := os.Getenv("INVENTORY_SERVICE_PORT")
	if grpcPort == "" {
		grpcPort = "8081"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterInventoryServiceServer(grpcServer, grpcProductHandler)

	reflection.Register(grpcServer)

	go func() {
		log.Printf("Inventory gRPC server started on port %s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	httpPort := os.Getenv("INVENTORY_SERVICE_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8082"
	}

	productHTTPHandler := NewProductHTTPHandler(productUseCase)

	router := mux.NewRouter()

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	router.HandleFunc("/api/products", productHTTPHandler.GetProducts).Methods("GET")
	router.HandleFunc("/api/products/{id}", productHTTPHandler.GetProduct).Methods("GET")
	router.HandleFunc("/api/products", productHTTPHandler.CreateProduct).Methods("POST")
	router.HandleFunc("/api/products/{id}", productHTTPHandler.UpdateProduct).Methods("PUT")
	router.HandleFunc("/api/products/{id}", productHTTPHandler.DeleteProduct).Methods("DELETE")

	router.HandleFunc("/api/products", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("OPTIONS")
	router.HandleFunc("/api/products/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("OPTIONS")

	httpServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: router,
	}

	go func() {
		log.Printf("Inventory HTTP server started on port %s", httpPort)
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

	log.Println("Product service gracefully stopped")
}
