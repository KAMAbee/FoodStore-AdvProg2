package grpc

import (
    "context"
    
    pb "AdvProg2/proto/product"
    "AdvProg2/usecase"
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
    product, err := h.productUseCase.CreateProduct(req.Name, req.Price, req.Stock)
    if err != nil {
        return nil, err
    }
    
    return &pb.Product{
        Id:    product.ID,
        Name:  product.Name,
        Price: product.Price,
        Stock: product.Stock,
    }, nil
}

func (h *ProductHandler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
    product, err := h.productUseCase.GetProduct(req.Id)
    if err != nil {
        return nil, err
    }
    
    return &pb.Product{
        Id:    product.ID,
        Name:  product.Name,
        Price: product.Price,
        Stock: product.Stock,
    }, nil
}

func (h *ProductHandler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.Product, error) {
    product, err := h.productUseCase.UpdateProduct(req.Id, req.Name, req.Price, req.Stock)
    if err != nil {
        return nil, err
    }
    
    return &pb.Product{
        Id:    product.ID,
        Name:  product.Name,
        Price: product.Price,
        Stock: product.Stock,
    }, nil
}

func (h *ProductHandler) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
    err := h.productUseCase.DeleteProduct(req.Id)
    if err != nil {
        return &pb.DeleteProductResponse{Success: false}, err
    }
    
    return &pb.DeleteProductResponse{Success: true}, nil
}

func (h *ProductHandler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
    products, total, err := h.productUseCase.ListProducts(req.Page, req.Limit)
    if err != nil {
        return nil, err
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