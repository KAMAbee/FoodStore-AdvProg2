package auth

import (
    "errors"
    "fmt"
    "os"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v4"
)

var (
    ErrInvalidToken = errors.New("invalid token")
    JWTSecret       []byte
)

func init() {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        secret = "123456"
    }
    JWTSecret = []byte(secret)
}

type Claims struct {
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

func GenerateToken(userID, username, role string) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour)
    
    claims := &Claims{
        UserID:   userID,
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(JWTSecret)
}

func ValidateToken(tokenString string) (*Claims, error) {
    tokenString = strings.TrimSpace(tokenString)
    
    claims := &Claims{}
    
    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return JWTSecret, nil
    })
    
    if err != nil {
        return nil, fmt.Errorf("error parsing token: %w", err)
    }
    
    if !token.Valid {
        return nil, ErrInvalidToken
    }
    
    return claims, nil
}