package auth

import (
    "errors"
    "fmt"
    "log"
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
    jwt.RegisteredClaims
}

func GenerateToken(userID, username string) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour)
    
    claims := &Claims{
        UserID:   userID,
        Username: username,
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
    
    if len(tokenString) > 10 {
        log.Printf("Token prefix: %s...", tokenString[:10])
    }
    
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