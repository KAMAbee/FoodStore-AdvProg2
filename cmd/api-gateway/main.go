package main

import (
    "log"
    "net/http"
    "os"
    "time"
    "github.com/gin-gonic/gin"

    "AdvProg2/middleware"
    "AdvProg2/pkg/cache"
    "bytes"
    "encoding/json"
    "io"
    "strings"
)

func proxyToService(serviceURL string, invalidateCache func(c *gin.Context, resp *http.Response) bool) gin.HandlerFunc {
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

        
        var bodyBytes []byte
        if invalidateCache != nil && (c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH") {
            bodyBytes, _ = io.ReadAll(c.Request.Body)
            c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
        }

        req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        
        if len(bodyBytes) > 0 {
            c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
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

        
        respBodyBytes, err := io.ReadAll(resp.Body)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading response: " + err.Error()})
            return
        }
        
        
        resp.Body = io.NopCloser(bytes.NewBuffer(respBodyBytes))

        
        c.Status(resp.StatusCode)
        for key, values := range resp.Header {
            for _, value := range values {
                c.Header(key, value)
            }
        }

        
        c.DataFromReader(resp.StatusCode, int64(len(respBodyBytes)), resp.Header.Get("Content-Type"), 
                         bytes.NewReader(respBodyBytes), nil)

        
        if invalidateCache != nil && (resp.StatusCode >= 200 && resp.StatusCode < 300) {
            if invalidateCache(c, resp) {
                log.Printf("Cache invalidated for %s %s", c.Request.Method, c.Request.URL.Path)
            }
        }
    }
}

func main() {
    // Настраиваем логирование в файл
    logFile, err := os.OpenFile("redis_cache.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        log.Printf("Ошибка при открытии файла логов: %v", err)
    } else {
        log.SetOutput(logFile)
        defer logFile.Close()
        log.Printf("Логирование настроено, логи сохраняются в redis_cache.log")
    }

    r := gin.New()
    r.Use(gin.Recovery())

    
    cacheClient := cache.New()
    log.Printf("Cache initialized")
    productCacheInvalidator := func(c *gin.Context, resp *http.Response) bool {
        method := c.Request.Method
        path := c.Request.URL.Path
        
        
        var productID string
        if method == "PUT" || method == "DELETE" || method == "PATCH" {
            parts := strings.Split(path, "/")
            if len(parts) > 3 {
                productID = parts[3]
                cacheKey := "product:" + productID
                cacheClient.Delete(cacheKey)
                log.Printf("Invalidated cache for product ID: %s", productID)
                return true
            }
        } else if method == "POST" && resp.StatusCode == http.StatusCreated {
            
            cacheClient.Delete("products:list")
            
            var productData map[string]interface{}
            body, err := io.ReadAll(resp.Body)
            resp.Body = io.NopCloser(bytes.NewBuffer(body)) 
            
            if err == nil && json.Unmarshal(body, &productData) == nil {
                if id, ok := productData["id"].(string); ok {
                    log.Printf("New product created with ID: %s", id)
                }
            }
            return true
        }
        
        return false
    }
    
    
    orderCacheInvalidator := func(c *gin.Context, resp *http.Response) bool {
        method := c.Request.Method
        path := c.Request.URL.Path
        
        if method == "POST" { 
            
            var orderData map[string]interface{}
            body, _ := io.ReadAll(c.Request.Body)
            c.Request.Body = io.NopCloser(bytes.NewBuffer(body)) 
            
            if err := json.Unmarshal(body, &orderData); err == nil {
                
                if userID, ok := orderData["user_id"].(string); ok {
                    cacheKey := "user:" + userID + ":orders"
                    cacheClient.Delete(cacheKey)
                    log.Printf("Invalidated cache for user orders: %s", userID)
                }
                
                
                if items, ok := orderData["items"].([]interface{}); ok {
                    for _, item := range items {
                        if itemMap, ok := item.(map[string]interface{}); ok {
                            if productID, ok := itemMap["product_id"].(string); ok {
                                cacheKey := "product:" + productID
                                cacheClient.Delete(cacheKey)
                                log.Printf("Invalidated cache for product ID: %s (order creation)", productID)
                            }
                        }
                    }
                }
            }
            return true
        } else if method == "PATCH" || method == "DELETE" {
            
            parts := strings.Split(path, "/")
            if len(parts) > 3 {
                orderID := parts[3]
                cacheKey := "order:" + orderID
                cacheClient.Delete(cacheKey)
                log.Printf("Invalidated cache for order ID: %s", orderID)
                
                
                userIDParam := c.Query("user_id")
                if userIDParam != "" {
                    userOrdersKey := "user:" + userIDParam + ":orders"
                    cacheClient.Delete(userOrdersKey)
                    log.Printf("Invalidated cache for user orders: %s", userIDParam)
                }
                
                
                if method == "PATCH" {
                    var statusUpdate map[string]interface{}
                    body, _ := io.ReadAll(c.Request.Body)
                    c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
                    
                    if json.Unmarshal(body, &statusUpdate) == nil {
                        if status, ok := statusUpdate["status"].(string); ok && (status == "completed" || status == "cancelled") {
                            cacheClient.Delete("products:list")
                            log.Printf("Order %s status changed to %s, invalidated products cache", orderID, status)
                        }
                    }
                }
                
                return true
            }
        }
        
        return false
    }
    
    
    userCacheInvalidator := func(c *gin.Context, resp *http.Response) bool {
        method := c.Request.Method
        path := c.Request.URL.Path
        
        if method == "PUT" || method == "PATCH" {
            parts := strings.Split(path, "/")
            if len(parts) > 3 {
                userID := parts[3]
                cacheKey := "user:" + userID
                cacheClient.Delete(cacheKey)
                log.Printf("Invalidated cache for user ID: %s", userID)
                return true
            }
        }
        
        return false
    }

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
        inventoryAPI.GET("", proxyToService(inventoryServiceURL, nil))
        inventoryAPI.GET("/:id", proxyToService(inventoryServiceURL, nil))
        inventoryAPI.POST("", proxyToService(inventoryServiceURL, productCacheInvalidator))
        inventoryAPI.PUT("/:id", proxyToService(inventoryServiceURL, productCacheInvalidator))
        inventoryAPI.DELETE("/:id", proxyToService(inventoryServiceURL, productCacheInvalidator))
    }

    orderServiceURL := os.Getenv("ORDER_SERVICE_URL")
    if orderServiceURL == "" {
        orderServiceURL = "http://localhost:8083"
    }

    orderAPI := r.Group("/api/orders")
    {
        orderAPI.GET("", proxyToService(orderServiceURL, nil))
        orderAPI.GET("/:id", proxyToService(orderServiceURL, nil))
        orderAPI.POST("", proxyToService(orderServiceURL, orderCacheInvalidator))
        orderAPI.PATCH("/:id", proxyToService(orderServiceURL, orderCacheInvalidator))
        orderAPI.DELETE("/:id", proxyToService(orderServiceURL, orderCacheInvalidator))
    }

    userServiceURL := os.Getenv("USER_SERVICE_URL")
    if userServiceURL == "" {
        userServiceURL = "http://localhost:8085"
    }

    userAPI := r.Group("/api/users")
    {
        userAPI.POST("/register", proxyToService(userServiceURL, nil))
        userAPI.POST("/login", proxyToService(userServiceURL, nil))
        userAPI.GET("/:id", proxyToService(userServiceURL, nil))
        userAPI.PUT("/:id", proxyToService(userServiceURL, userCacheInvalidator))
        userAPI.PATCH("/:id", proxyToService(userServiceURL, userCacheInvalidator))
    }

    
    adminServiceURL := os.Getenv("ADMIN_SERVICE_URL")
    if adminServiceURL == "" {
        adminServiceURL = "http://localhost:8085"
    }

    adminAPI := r.Group("/api/admin")
    adminAPI.Use(middleware.AdminRequired())
    {
        adminAPI.POST("/products", proxyToService(adminServiceURL, productCacheInvalidator))
        adminAPI.PUT("/products/:id", proxyToService(adminServiceURL, productCacheInvalidator))
        adminAPI.DELETE("/products/:id", proxyToService(adminServiceURL, productCacheInvalidator))
    }

    
    if os.Getenv("ENV") != "production" {
        r.GET("/api/debug/cache-stats", func(c *gin.Context) {
            stats := cacheClient.GetStats()
            c.JSON(http.StatusOK, stats)
        })
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