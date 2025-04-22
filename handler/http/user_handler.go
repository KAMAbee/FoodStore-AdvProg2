package grpc

import (
    "encoding/json"
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
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    user, err := h.userUseCase.Register(req.Username, req.Password)
    if err != nil {
        if err == repository.ErrUsernameAlreadyExists {
            http.Error(w, "Username already exists", http.StatusConflict)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

func (h *UserHTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    user, err := h.userUseCase.Login(req.Username, req.Password)
    if err != nil {
        if err == repository.ErrInvalidCredentials {
            http.Error(w, "Invalid credentials", http.StatusUnauthorized)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(user)
}

func (h *UserHTTPHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    vars := mux.Vars(r)
    userID := vars["id"]

    user, err := h.userUseCase.GetProfile(userID)
    if err != nil {
        if err == repository.ErrUserNotFound {
            http.Error(w, "User not found", http.StatusNotFound)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(user)
}