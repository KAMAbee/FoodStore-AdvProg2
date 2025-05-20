# FoodStore-AdvProg2

FoodStore is a microservices-based platform built with Go, designed for managing products, orders, users, and administrative tasks.  
It leverages Redis for caching, NATS for messaging, PostgreSQL for data persistence, and an API Gateway to route requests.  

---

## Implemented features 
- Clean Architecture  
- gRPC
- Message Queues
- Databases and Caches (including migrations and transactions)
- Sending Emails
- Testing
- Web frontend (HTML/SCSS/JS)

##  Prerequisites

- **Go**: Version 1.20 or higher  
- **Docker**: For running Redis and NATS  
- **PostgreSQL**: Version 11 or higher  
- **Gmail Account**: For SMTP email sending (with App Password)  
- **NATS**: For messaging between services  

---

##  Setup

### 1. Clone the Repository

```bash
git clone https://github.com/KAMAbee/FoodStore-AdvProg2
cd FoodStore-AdvProg2
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Set Up Environment Variables

Create a `.env` file in the project root:

```dotenv
# Database (PostgreSQL)
DB=URLOnYourDB

# Service Ports and URLs
API_GATEWAY_PORT=8080
INVENTORY_SERVICE_PORT=8081
INVENTORY_SERVICE_HTTP_PORT=8082
INVENTORY_SERVICE_URL=http://localhost:8082
ORDER_SERVICE_PORT=8083
ORDER_SERVICE_HTTP_PORT=8093
ORDER_SERVICE_URL=http://localhost:8093
USER_SERVICE_PORT=8084
USER_SERVICE_HTTP_PORT=8085
USER_SERVICE_URL=http://localhost:8085
EMAIL_SERVICE_PORT=8086
EMAIL_SERVICE_URL=http://localhost:8086

# JWT Secret
JWT_SECRET=123456

# NATS
NATS_URL=nats://localhost:4222

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

# SMTP Configuration (Gmail)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=youremail
SMTP_PASSWORD=yourappPassword
```

> **Notes**:  
> - Ensure SMTP credentials are valid (App Password with 2-Step Verification).  
> - Update `DB` if your PostgreSQL setup differs.  
> - Use a secure, random `JWT_SECRET` in production.

### 4. Set Up PostgreSQL

Set up PostgreSQL locally

### 5. Run Dependencies

#### Redis

```bash
docker run --name redis -d -p 6379:6379 redis
docker exec -it redis redis-cli
```

#### NATS

```bash
docker run --name nats -d -p 4222:4222 -p 8222:8222 nats
```

### 6. Run Database Migrations

```bash
go run cmd/migration/main.go
```

### 7. Run Microservices

```bash
# API Gateway
go run cmd/api-gateway/main.go

# Product Service
go run cmd/product-service/main.go

# Order Service
go run cmd/order-service/main.go

# User Service
go run cmd/user-service/main.go

# Consumer Services
go run cmd/consumer-service/main.go
go run cmd/admin-consumer/main.go

# Email Service
go run cmd/email-service/main.go
```

### 8. Access the Application

- **Admin Panel**: [http://localhost:8080/admin](http://localhost:8080/admin)  
- **API Gateway**: [http://localhost:8080/api/](http://localhost:8080/api/)  
- **NATS Monitoring**: [http://localhost:8222](http://localhost:8222)  
- **Redis CLI**: Use `docker exec -it redis redis-cli`

---

##  Usage

### Admin Panel

- **Login/Register** at `/login` or `/register`
- **Manage Products** at `/admin`
- **Send Emails** (admin-only) at `/admin`
- **Order products** at `/orders`

### GRPC Endpoints

- Product Service (Inventory Service):
```
CreateProduct - create a new product
GetProduct - get product by ID
UpdateProduct - update existing product
DeleteProduct - delete product
ListProducts - get list of products with pagination
SearchProducts - search products by filters
```

- Order Service (Order Service):
```
CreateOrder - create new order
GetOrder - get order by ID
GetUserOrders - get all user orders
UpdateOrderStatus - update order request
CancelOrder - cancel order
```

- User Service (User Service):
```
Registration - register new user
Login - user authorization
GetUser - get user profile
```

#### Example: Send an Email

```bash
curl -X POST http://localhost:8080/api/email/send \
  -H "Content-Type: application/json" \
  -H "X-User-Role: admin" \
  -d '{"to":"test@example.com","subject":"Test Email","body":"Hello from FoodStore!"}'
```

### Redis Interaction

```bash
docker exec -it redis redis-cli
# Commands:
# KEYS *         -> List keys
# GET <key>      -> Get value
# MONITOR        -> Real-time view
```

### Run tests
- Unit test
```bash
go test -v ./infrastructure/db -run TestPostgresOrderRepository_Create
```

- Integration test: 
```bash
go test -v ./tests/integration -run TestOrderCreationFlow
```
### Run database migration
```bash
migrate -path ./migrations -database "YOUR-DB-PATH" up
```



