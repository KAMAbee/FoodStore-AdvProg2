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
    
    db "AdvProg2/infrastructure/db"
    grpcHandler "AdvProg2/handler/grpc"
    "AdvProg2/domain"
    pb "AdvProg2/proto/product"
    "AdvProg2/usecase"
    
    "github.com/gorilla/mux"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
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
        "total": total,
        "page": page,
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
    dbConn, err := db.NewPostgresConnection()
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer dbConn.Close()
    
    productRepo := db.NewPostgresProductRepository(dbConn)
    productUseCase := usecase.NewProductUseCase(productRepo)
    
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
    router.HandleFunc("/api/products", productHTTPHandler.GetProducts).Methods("OPTIONS")
    router.HandleFunc("/api/products/{id}", productHTTPHandler.GetProduct).Methods("OPTIONS")
    
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
}