package grpc

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/gorilla/mux"
    
    "AdvProg2/repository"
    "AdvProg2/usecase"
)

type UserHTTPHandler struct {
    userUseCase *usecase.UserUseCase
}

func NewUserHTTPHandler(userUseCase *usecase.UserUseCase) *UserHTTPHandler {
    return &UserHTTPHandler{
        userUseCase: userUseCase,
    }
}

func (h *UserHTTPHandler) Register(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
        Role     string `json:"role"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("Error decoding register request: %v", err)
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    log.Printf("Register request for username: %s with role: %s", req.Username, req.Role)

    if req.Role == "" {
        req.Role = "user"
    }

    authResponse, err := h.userUseCase.Register(req.Username, req.Password, req.Role)
    if err != nil {
        log.Printf("Register error: %v", err)
        if err == repository.ErrUsernameAlreadyExists {
            http.Error(w, "Username already exists", http.StatusConflict)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    log.Printf("User registered successfully: %s with role: %s", req.Username, authResponse.User.Role)

    http.SetCookie(w, &http.Cookie{
        Name:     "auth_token",
        Value:    authResponse.Token,
        Path:     "/",
        HttpOnly: false,
        MaxAge:   86400,
        SameSite: http.SameSiteLaxMode,
    })

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(authResponse)
}

func (h *UserHTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("Error decoding login request: %v", err)
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    log.Printf("Login request for username: %s", req.Username)

    authResponse, err := h.userUseCase.Login(req.Username, req.Password)
    if err != nil {
        log.Printf("Login error: %v", err)
        if err == repository.ErrInvalidCredentials {
            http.Error(w, "Invalid credentials", http.StatusUnauthorized)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    log.Printf("User logged in successfully: %s with role: %s", req.Username, authResponse.User.Role)

    http.SetCookie(w, &http.Cookie{
        Name:     "auth_token",
        Value:    authResponse.Token,
        Path:     "/",
        HttpOnly: false, 
        MaxAge:   86400,
        SameSite: http.SameSiteLaxMode,
    })

    json.NewEncoder(w).Encode(authResponse)
}

func (h *UserHTTPHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    vars := mux.Vars(r)
    userID := vars["id"]

    log.Printf("GetProfile request for user ID: %s", userID)

    user, err := h.userUseCase.GetProfile(userID)
    if err != nil {
        log.Printf("GetProfile error: %v", err)
        if err == repository.ErrUserNotFound {
            http.Error(w, "User not found", http.StatusNotFound)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(user)
}