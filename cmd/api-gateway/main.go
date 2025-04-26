package main

import (
    "log"
    "net/http"
    "os"
    "time"
    "github.com/gin-gonic/gin"

    "AdvProg2/middleware"
)

func proxyToService(serviceURL string) gin.HandlerFunc {
    return func(c *gin.Context) {
        if serviceURL == "" {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Service URL not configured"})
            return
        }

        client := &http.Client{Timeout: 10 * time.Second}
        targetURL := serviceURL + c.Request.URL.Path
        if c.Request.URL.RawQuery != "" {
            targetURL += "?" + c.Request.URL.RawQuery
        }

        req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        for key, values := range c.Request.Header {
            for _, value := range values {
                req.Header.Add(key, value)
            }
        }

        if requestID, exists := c.Get("RequestID"); exists {
            req.Header.Set("X-Request-ID", requestID.(string))
        }

        resp, err := client.Do(req)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Service unavailable: " + err.Error()})
            return
        }
        defer resp.Body.Close()

        c.Status(resp.StatusCode)
        for key, values := range resp.Header {
            for _, value := range values {
                c.Header(key, value)
            }
        }

        c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
    }
}

func main() {
    r := gin.New()
    r.Use(gin.Recovery())

    r.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    })

    r.Static("/static", "./public")
    r.LoadHTMLGlob("public/*.html")

    r.GET("/login", func(c *gin.Context) {
        c.HTML(http.StatusOK, "login.html", nil)
    })
    r.GET("/register", func(c *gin.Context) {
        c.HTML(http.StatusOK, "register.html", nil)
    })

    r.Use(middleware.AuthMiddleware())

    r.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "order.html", nil)
    })
    r.GET("/admin", func(c *gin.Context) {
        c.HTML(http.StatusOK, "admin.html", nil)
    })
    r.GET("/order", func(c *gin.Context) {
        c.HTML(http.StatusOK, "order.html", nil)
    })
    r.GET("/profile", func(c *gin.Context) {
        c.HTML(http.StatusOK, "profile.html", nil)
    })

    inventoryServiceURL := os.Getenv("INVENTORY_SERVICE_URL")
    if inventoryServiceURL == "" {
        inventoryServiceURL = "http://localhost:8082"
    }

    r.GET("/api/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "up",
            "time":   time.Now().Format(time.RFC3339),
        })
    })

    inventoryAPI := r.Group("/api/products")
    {
        inventoryAPI.GET("", proxyToService(inventoryServiceURL))
        inventoryAPI.GET("/:id", proxyToService(inventoryServiceURL))
        inventoryAPI.POST("", proxyToService(inventoryServiceURL))
        inventoryAPI.PUT("/:id", proxyToService(inventoryServiceURL))
        inventoryAPI.DELETE("/:id", proxyToService(inventoryServiceURL))
    }

    orderServiceURL := os.Getenv("ORDER_SERVICE_URL")
    if orderServiceURL == "" {
        orderServiceURL = "http://localhost:8083"
    }

    orderAPI := r.Group("/api/orders")
    {
        orderAPI.GET("", proxyToService(orderServiceURL))
        orderAPI.GET("/:id", proxyToService(orderServiceURL))
        orderAPI.POST("", proxyToService(orderServiceURL))
        orderAPI.PATCH("/:id", proxyToService(orderServiceURL))
        orderAPI.DELETE("/:id", proxyToService(orderServiceURL))
    }

    userServiceURL := os.Getenv("USER_SERVICE_URL")
    if userServiceURL == "" {
        userServiceURL = "http://localhost:8085"
    }

    userAPI := r.Group("/api/users")
    {
        userAPI.POST("/register", proxyToService(userServiceURL))
        userAPI.POST("/login", proxyToService(userServiceURL))
        userAPI.GET("/:id", proxyToService(userServiceURL))
    }

    port := os.Getenv("API_GATEWAY_PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("API Gateway starting on port %s", port)
    if err := r.Run(":" + port); err != nil {
        log.Fatalf("Failed to start API Gateway: %v", err)
    }
}