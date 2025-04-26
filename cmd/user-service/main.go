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
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"

    "AdvProg2/infrastructure/db"
    grpcHandler "AdvProg2/handler/grpc"
    httpHandler "AdvProg2/handler/http"
    pb "AdvProg2/proto/user"
    "AdvProg2/usecase"
)

func main() {
    godotenv.Load()

    dbConn, err := db.NewPostgresConnection()
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer dbConn.Close()

    userRepo, err := db.NewPostgresUserRepository(dbConn)
    if err != nil {
        log.Fatalf("Failed to create user repository: %v", err)
    }
    
    userUseCase := usecase.NewUserUseCase(userRepo)

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

    router := mux.NewRouter()

    userHTTPHandler := httpHandler.NewUserHTTPHandler(userUseCase)

    router.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin", "*")
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }

            next.ServeHTTP(w, r)
        })
    })

    router.HandleFunc("/api/users/register", userHTTPHandler.Register).Methods("POST")
    router.HandleFunc("/api/users/login", userHTTPHandler.Login).Methods("POST")
    router.HandleFunc("/api/users/{id}", userHTTPHandler.GetProfile).Methods("GET")

    router.HandleFunc("/api/users/register", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }).Methods("OPTIONS")
    router.HandleFunc("/api/users/login", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }).Methods("OPTIONS")
    router.HandleFunc("/api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }).Methods("OPTIONS")

    httpPort := os.Getenv("USER_SERVICE_HTTP_PORT")
    if httpPort == "" {
        httpPort = "8085"
    }

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
}