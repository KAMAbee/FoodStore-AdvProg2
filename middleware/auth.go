package middleware

import (
    "log"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"

    "AdvProg2/pkg/auth"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        path := c.Request.URL.Path
        log.Printf("Auth middleware checking: %s", path)

        if path == "/login" || 
           path == "/register" || 
           path == "/api/users/login" || 
           path == "/api/users/register" ||
           strings.HasPrefix(path, "/static/") {
            log.Printf("Public route accessed: %s", path)
            c.Next()
            return
        }

        var tokenString string
        
        authHeader := c.GetHeader("Authorization")
        if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
            tokenString = strings.TrimPrefix(authHeader, "Bearer ")
            log.Printf("Found token in Authorization header")
        }
        
        if tokenString == "" {
            cookie, err := c.Request.Cookie("auth_token")
            if err == nil && cookie.Value != "" {
                tokenString = cookie.Value
                log.Printf("Found token in cookie")
            }
        }

        if tokenString == "" {
            log.Printf("No token found for path: %s", path)
            
            if strings.HasPrefix(path, "/api/") {
                c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
            } else {
                c.Redirect(http.StatusFound, "/login")
                c.Abort()
            }
            return
        }

        claims, err := auth.ValidateToken(tokenString)
        if err != nil {
            log.Printf("Token validation failed: %v", err)
            
            c.SetCookie("auth_token", "", -1, "/", "", false, true)
            
            if strings.HasPrefix(path, "/api/") {
                c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            } else {
                c.Redirect(http.StatusFound, "/login")
                c.Abort()
            }
            return
        }

        if strings.HasPrefix(path, "/admin") || strings.HasPrefix(path, "/api/admin") {
            if claims.Role != "admin" {
                log.Printf("Access denied: user %s with role %s trying to access admin route %s", 
                    claims.Username, claims.Role, path)
                
                if strings.HasPrefix(path, "/api/") {
                    c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
                } else {
                    c.Redirect(http.StatusFound, "/profile")
                    c.Abort()
                }
                return
            }
        }

        log.Printf("Token valid for user: %s (ID: %s, Role: %s)", 
            claims.Username, claims.UserID, claims.Role)
        c.Set("userID", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("userRole", claims.Role)
        
        c.Next()
    }
}

func AdminRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("userRole")
        if !exists || role != "admin" {
            if strings.HasPrefix(c.Request.URL.Path, "/api/") {
                c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
            } else {
                c.Redirect(http.StatusFound, "/profile")
                c.Abort()
            }
            return
        }
        c.Next()
    }
}