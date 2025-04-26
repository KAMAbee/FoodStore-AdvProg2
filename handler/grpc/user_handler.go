package grpc

import (
    "context"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    
    "AdvProg2/repository"
    pb "AdvProg2/proto/user"
    "AdvProg2/usecase"
)

type UserHandler struct {
    pb.UnimplementedUserServiceServer
    userUseCase *usecase.UserUseCase
}

func NewUserHandler(userUseCase *usecase.UserUseCase) *UserHandler {
    return &UserHandler{
        userUseCase: userUseCase,
    }
}

func (h *UserHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.UserResponse, error) {
    if req.Username == "" || req.Password == "" {
        return nil, status.Error(codes.InvalidArgument, "username and password are required")
    }

    authResponse, err := h.userUseCase.Register(req.Username, req.Password, req.Role)
    if err != nil {
        if err == repository.ErrUsernameAlreadyExists {
            return nil, status.Error(codes.AlreadyExists, "username already exists")
        }
        return nil, status.Error(codes.Internal, err.Error())
    }

    return &pb.UserResponse{
        Id:       authResponse.User.ID,
        Username: authResponse.User.Username,
        Token:    authResponse.Token,
        Role:     authResponse.User.Role,
    }, nil
}

func (h *UserHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.UserResponse, error) {
    if req.Username == "" || req.Password == "" {
        return nil, status.Error(codes.InvalidArgument, "username and password are required")
    }

    authResponse, err := h.userUseCase.Login(req.Username, req.Password)
    if err != nil {
        if err == repository.ErrInvalidCredentials {
            return nil, status.Error(codes.Unauthenticated, "invalid credentials")
        }
        return nil, status.Error(codes.Internal, err.Error())
    }

    return &pb.UserResponse{
        Id:       authResponse.User.ID,
        Username: authResponse.User.Username,
        Token:    authResponse.Token,
        Role:     authResponse.User.Role,
    }, nil
}

func (h *UserHandler) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.UserResponse, error) {
    if req.Id == "" {
        return nil, status.Error(codes.InvalidArgument, "user ID is required")
    }

    user, err := h.userUseCase.GetProfile(req.Id)
    if err != nil {
        if err == repository.ErrUserNotFound {
            return nil, status.Error(codes.NotFound, "user not found")
        }
        return nil, status.Error(codes.Internal, err.Error())
    }

    return &pb.UserResponse{
        Id:       user.ID,
        Username: user.Username,
        Role:     user.Role,
    }, nil
}