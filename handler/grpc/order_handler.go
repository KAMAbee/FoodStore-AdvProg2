package grpc

import (
    "context"
    "time"
    
    pb "AdvProg2/proto/order"
    "AdvProg2/usecase"
    "AdvProg2/domain"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type OrderHandler struct {
    pb.UnimplementedOrderServiceServer
    orderUseCase *usecase.OrderUseCase
}

func NewOrderHandler(orderUseCase *usecase.OrderUseCase) *OrderHandler {
    return &OrderHandler{
        orderUseCase: orderUseCase,
    }
}

// Конвертация домена Order в gRPC сообщение Order
func domainOrderToProto(order *domain.Order) *pb.Order {
    if order == nil {
        return nil
    }
    
    protoOrder := &pb.Order{
        Id:         order.ID,
        UserId:     order.UserID,
        Status:     order.Status,
        TotalPrice: order.TotalPrice,
        CreatedAt:  order.CreatedAt.Format(time.RFC3339),
        Items:      make([]*pb.OrderItem, 0, len(order.Items)),
    }
    
    for _, item := range order.Items {
        protoItem := &pb.OrderItem{
            Id:        item.ID,
            OrderId:   item.OrderID,
            ProductId: item.ProductID,
            Quantity:  item.Quantity,
            Price:     item.Price,
        }
        
        if item.Product != nil {
            protoItem.Product = &pb.Product{
                Id:    item.Product.ID,
                Name:  item.Product.Name,
                Price: item.Product.Price,
                Stock: item.Product.Stock,
            }
        }
        
        protoOrder.Items = append(protoOrder.Items, protoItem)
    }
    
    return protoOrder
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.Order, error) {
    if req.UserId == "" {
        return nil, status.Error(codes.InvalidArgument, "user ID is required")
    }
    
    if len(req.Items) == 0 {
        return nil, status.Error(codes.InvalidArgument, "order must have at least one item")
    }
    
    var orderItems []struct{ProductID string; Quantity int32}
    
    for _, item := range req.Items {
        orderItems = append(orderItems, struct{ProductID string; Quantity int32}{
            ProductID: item.ProductId,
            Quantity:  item.Quantity,
        })
    }
    
    order, err := h.orderUseCase.CreateOrder(req.UserId, orderItems)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    
    return domainOrderToProto(order), nil
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.Order, error) {
    if req.Id == "" {
        return nil, status.Error(codes.InvalidArgument, "order ID is required")
    }
    
    order, err := h.orderUseCase.GetOrder(req.Id)
    if err != nil {
        return nil, status.Error(codes.NotFound, err.Error())
    }
    
    return domainOrderToProto(order), nil
}

func (h *OrderHandler) GetUserOrders(ctx context.Context, req *pb.GetUserOrdersRequest) (*pb.ListOrdersResponse, error) {
    if req.UserId == "" {
        return nil, status.Error(codes.InvalidArgument, "user ID is required")
    }
    
    orders, total, err := h.orderUseCase.GetUserOrders(req.UserId, req.Page, req.Limit)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    
    protoOrders := make([]*pb.Order, 0, len(orders))
    for _, order := range orders {
        protoOrders = append(protoOrders, domainOrderToProto(order))
    }
    
    return &pb.ListOrdersResponse{
        Orders: protoOrders,
        Total:  total,
    }, nil
}

func (h *OrderHandler) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.Order, error) {
    if req.Id == "" {
        return nil, status.Error(codes.InvalidArgument, "order ID is required")
    }
    
    if req.Status == "" {
        return nil, status.Error(codes.InvalidArgument, "status is required")
    }
    
    order, err := h.orderUseCase.UpdateOrderStatus(req.Id, req.Status)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    
    return domainOrderToProto(order), nil
}

func (h *OrderHandler) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
    if req.Id == "" {
        return nil, status.Error(codes.InvalidArgument, "order ID is required")
    }
    
    err := h.orderUseCase.CancelOrder(req.Id)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    
    return &pb.CancelOrderResponse{
        Success: true,
    }, nil
}