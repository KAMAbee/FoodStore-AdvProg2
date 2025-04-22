package grpc

import (
    "encoding/json"
    "net/http"
    "strconv"
    
    "github.com/gorilla/mux"
    "AdvProg2/usecase"
)

type OrderHTTPHandler struct {
    orderUseCase *usecase.OrderUseCase
}

func NewOrderHTTPHandler(orderUseCase *usecase.OrderUseCase) *OrderHTTPHandler {
    return &OrderHTTPHandler{
        orderUseCase: orderUseCase,
    }
}

func (h *OrderHTTPHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    type OrderItemRequest struct {
        ProductID string `json:"product_id"`
        Quantity  int32  `json:"quantity"`
    }
    
    type CreateOrderRequest struct {
        UserID string            `json:"user_id"`
        Items  []OrderItemRequest `json:"items"`
    }
    
    var req CreateOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    if req.UserID == "" {
        http.Error(w, "User ID is required", http.StatusBadRequest)
        return
    }
    
    if len(req.Items) == 0 {
        http.Error(w, "Order must have at least one item", http.StatusBadRequest)
        return
    }
    
    var orderItems []struct{ProductID string; Quantity int32}
    for _, item := range req.Items {
        orderItems = append(orderItems, struct{ProductID string; Quantity int32}{
            ProductID: item.ProductID,
            Quantity:  item.Quantity,
        })
    }
    
    order, err := h.orderUseCase.CreateOrder(req.UserID, orderItems)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "order_id": order.ID,
        "status":   order.Status,
        "message":  "Order created successfully",
    })
}

func (h *OrderHTTPHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    vars := mux.Vars(r)
    id := vars["id"]
    
    order, err := h.orderUseCase.GetOrder(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    
    json.NewEncoder(w).Encode(order)
}

func (h *OrderHTTPHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    userID := r.URL.Query().Get("user_id")
    if userID == "" {
        http.Error(w, "User ID is required", http.StatusBadRequest)
        return
    }
    
    page := int32(1)
    limit := int32(10)
    
    if pageStr := r.URL.Query().Get("page"); pageStr != "" {
        if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
            page = int32(p)
        }
    }
    
    if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
        if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
            limit = int32(l)
        }
    }
    
    orders, total, err := h.orderUseCase.GetUserOrders(userID, page, limit)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    response := map[string]interface{}{
        "orders": orders,
        "total":  total,
        "page":   page,
        "limit":  limit,
    }
    
    json.NewEncoder(w).Encode(response)
}

func (h *OrderHTTPHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    vars := mux.Vars(r)
    id := vars["id"]
    
    type UpdateStatusRequest struct {
        Status string `json:"status"`
    }
    
    var req UpdateStatusRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    if req.Status == "" {
        http.Error(w, "Status is required", http.StatusBadRequest)
        return
    }
    
    order, err := h.orderUseCase.UpdateOrderStatus(id, req.Status)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(order)
}

func (h *OrderHTTPHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    
    err := h.orderUseCase.CancelOrder(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Order cancelled successfully",
    })
}