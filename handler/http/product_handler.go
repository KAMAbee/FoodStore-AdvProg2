package grpc

import (
    "context"
    
    pb "AdvProg2/proto/product"
    "AdvProg2/usecase"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type ProductHandler struct {
    pb.UnimplementedInventoryServiceServer
    productUseCase *usecase.ProductUseCase
}

func NewProductHandler(productUseCase *usecase.ProductUseCase) *ProductHandler {
    return &ProductHandler{
        productUseCase: productUseCase,
    }
}

func (h *ProductHandler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.Product, error) {
    if req.Name == "" {
        return nil, status.Error(codes.InvalidArgument, "product name is required")
    }
    if req.Price < 0 {
        return nil, status.Error(codes.InvalidArgument, "product price cannot be negative")
    }
    if req.Stock < 0 {
        return nil, status.Error(codes.InvalidArgument, "product stock cannot be negative")
    }
    
    product, err := h.productUseCase.CreateProduct(req.Name, req.Price, req.Stock)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    
    return &pb.Product{
        Id:    product.ID,
        Name:  product.Name,
        Price: product.Price,
        Stock: product.Stock,
    }, nil
}

func (h *ProductHandler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
    if req.Id == "" {
        return nil, status.Error(codes.InvalidArgument, "product ID is required")
    }
    
    product, err := h.productUseCase.GetProduct(req.Id)
    if err != nil {
        return nil, status.Error(codes.NotFound, err.Error())
    }
    
    return &pb.Product{
        Id:    product.ID,
        Name:  product.Name,
        Price: product.Price,
        Stock: product.Stock,
    }, nil
}

func (h *ProductHandler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.Product, error) {
    if req.Id == "" {
        return nil, status.Error(codes.InvalidArgument, "product ID is required")
    }
    
    product, err := h.productUseCase.UpdateProduct(req.Id, req.Name, req.Price, req.Stock)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    
    return &pb.Product{
        Id:    product.ID,
        Name:  product.Name,
        Price: product.Price,
        Stock: product.Stock,
    }, nil
}

func (h *ProductHandler) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
    if req.Id == "" {
        return nil, status.Error(codes.InvalidArgument, "product ID is required")
    }
    
    err := h.productUseCase.DeleteProduct(req.Id)
    if err != nil {
        return nil, status.Error(codes.NotFound, err.Error())
    }
    
    return &pb.DeleteProductResponse{Success: true}, nil
}

func (h *ProductHandler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
    products, total, err := h.productUseCase.ListProducts(req.Page, req.Limit)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    
    pbProducts := make([]*pb.Product, len(products))
    for i, product := range products {
        pbProducts[i] = &pb.Product{
            Id:    product.ID,
            Name:  product.Name,
            Price: product.Price,
            Stock: product.Stock,
        }
    }
    
    return &pb.ListProductsResponse{
        Products: pbProducts,
        Total:    total,
    }, nil
}