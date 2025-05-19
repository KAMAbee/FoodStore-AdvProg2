FoodStore-AdvProg2
FoodStore is a microservices-based e-commerce platform built with Go, designed for managing products, orders, users, and administrative tasks. It leverages Redis for caching, NATS for messaging, PostgreSQL for data persistence, and an API Gateway to route requests. The admin panel supports product management and email notifications.
Prerequisites

Go: Version 1.20 or higher
Docker: For running Redis and NATS
PostgreSQL: Version 11 or higher
Gmail Account: For SMTP email sending (with App Password)
NATS: For messaging between services

Setup
1. Clone the Repository
git clone <repository-url>
cd FoodStore-AdvProg2

2. Install Dependencies
Install Go dependencies:
go mod download

3. Set Up Environment Variables
Create a .env file in the project root with the following configuration:
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

Note: 

Ensure the SMTP_USERNAME and SMTP_PASSWORD are valid. The provided SMTP_PASSWORD is a Gmail App Password, which requires 2-Step Verification enabled in your Google Account.
Update the DB connection string if your PostgreSQL setup differs (e.g., different host, port, user, or password).
The JWT_SECRET is set to a simple value for development; use a secure, random string in production.

4. Set Up PostgreSQL
Install and run PostgreSQL locally or in a Docker container.
Local PostgreSQL

Install PostgreSQL: https://www.postgresql.org/download/
Create a database named postgres:psql -U postgres -c "CREATE DATABASE postgres;"


Verify the connection:psql "postgresql://postgres:12345678@localhost:5432/postgres?sslmode=disable"



Docker PostgreSQL
Run PostgreSQL in Docker:
docker run --name postgres -d -p 5432:5432 -e POSTGRES_PASSWORD=12345678 postgres

5. Run Dependencies
Redis
Start a Redis container for caching:
docker run --name redis -d -p 6379:6379 redis

Verify Redis is running:
docker exec -it redis redis-cli

Run PING in the Redis CLI (expected output: PONG).
NATS
Start a NATS server for messaging:
docker run --name nats -d -p 4222:4222 -p 8222:8222 nats

Verify NATS is running:
curl http://localhost:8222

6. Run Database Migrations
Run the migration service to set up the PostgreSQL database schema:
go run cmd/migration/main.go

7. Run Microservices
Start each service in a separate terminal or use a process manager (e.g., tmux, screen). Ensure the .env file is loaded or environment variables are set.
# API Gateway
go run cmd/api-gateway/main.go

# Product Service (Inventory)
go run cmd/product-service/main.go

# Order Service
go run cmd/order-service/main.go

# User Service
go run cmd/user-service/main.go

# Consumer Service
go run cmd/consumer-service/main.go

# Admin Consumer
go run cmd/admin-consumer/main.go

# Email Service
go run cmd/email-service/main.go

8. Access the Application

Admin Panel: Open http://localhost:8080/admin in a browser (requires admin login).
API Gateway: API endpoints are available at http://localhost:8080/api/.
NATS Monitoring: Access at http://localhost:8222.
Redis CLI: Connect with docker exec -it redis redis-cli.

Usage
Admin Panel

Login: Use /login or /register to create an admin account.
Manage Products: Add, edit, or delete products via /admin.
Send Emails: Use the "Send Email" form to send notifications (admin-only).

API Endpoints

Products: GET/POST/PUT/DELETE /api/products
Orders: GET/POST/PATCH/DELETE /api/orders
Users: GET/POST/PUT/PATCH /api/users
Email: POST /api/email/send (admin-only)
Health Check: GET /api/health
Cache Stats: GET /api/debug/cache-stats (non-production)

Example API Request
Send an email:
curl -X POST http://localhost:8080/api/email/send \
  -H "Content-Type: application/json" \
  -H "X-User-Role: admin" \
  -d '{"to":"test@example.com","subject":"Test Email","body":"Hello from FoodStore!"}'

Redis Interaction
Inspect the cache:
docker exec -it redis redis-cli

Commands:

KEYS *: List all keys (e.g., product:<uuid>).
GET <key>: Retrieve a key’s value.
MONITOR: Watch real-time operations.

Docker Compose (Optional)
For easier management, use this docker-compose.yml:
version: '3.8'
services:
  redis:
    image: redis:latest
    container_name: redis
    ports:
      - "6379:6379"
    restart: unless-stopped
  nats:
    image: nats:latest
    container_name: nats
    ports:
      - "4222:4222"
      - "8222:8222"
    restart: unless-stopped
  postgres:
    image: postgres:latest
    container_name: postgres
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_PASSWORD=12345678
      - POSTGRES_DB=postgres
    restart: unless-stopped
  api-gateway:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB=URLOnYourDB
      - API_GATEWAY_PORT=8080
      - INVENTORY_SERVICE_URL=http://product:8082
      - ORDER_SERVICE_URL=http://order:8093
      - USER_SERVICE_URL=http://user:8085
      - EMAIL_SERVICE_URL=http://email:8086
      - JWT_SECRET=123456
      - NATS_URL=nats://nats:4222
      - REDIS_ADDR=redis:6379
      - REDIS_PASSWORD=
    depends_on:
      - redis
      - nats
      - postgres
  email:
    build:
      context: .
      dockerfile: cmd/email-service/Dockerfile
    ports:
      - "8086:8086"
    environment:
      - SMTP_HOST=smtp.gmail.com
      - SMTP_PORT=587
      - SMTP_USERNAME=olzhas200696@gmail.com
      - SMTP_PASSWORD=fdfy ubqq tqps dvcd
      - EMAIL_SERVICE_PORT=8086
    depends_on:
      - redis
      - nats
  product:
    build: .
    ports:
      - "8081:8081"
      - "8082:8082"
    environment:
      - DB=postgresql://postgres:12345678@postgres:5432/postgres?sslmode=disable
      - INVENTORY_SERVICE_PORT=8081
      - INVENTORY_SERVICE_HTTP_PORT=8082
      - NATS_URL=nats://nats:4222
      - REDIS_ADDR=redis:6379
    depends_on:
      - redis
      - nats
      - postgres
  order:
    build: .
    ports:
      - "8083:8083"
      - "8093:8093"
    environment:
      - DB=postgresql://postgres:12345678@postgres:5432/postgres?sslmode=disable
      - ORDER_SERVICE_PORT=8083
      - ORDER_SERVICE_HTTP_PORT=8093
      - NATS_URL=nats://nats:4222
      - REDIS_ADDR=redis:6379
    depends_on:
      - redis
      - nats
      - postgres
  user:
    build: .
    ports:
      - "8084:8084"
      - "8085:8085"
    environment:
      - DB=postgresql://postgres:12345678@postgres:5432/postgres?sslmode=disable
      - USER_SERVICE_PORT=8084
      - USER_SERVICE_HTTP_PORT=8085
      - JWT_SECRET=123456
      - NATS_URL=nats://nats:4222
      - REDIS_ADDR=redis:6379
    depends_on:
      - redis
      - nats
      - postgres
  consumer:
    build: .
    environment:
      - DB=postgresql://postgres:12345678@postgres:5432/postgres?sslmode=disable
      - NATS_URL=nats://nats:4222
      - REDIS_ADDR=redis:6379
    depends_on:
      - redis
      - nats
      - postgres
  admin-consumer:
    build: .
    environment:
      - DB=postgresql://postgres:12345678@postgres:5432/postgres?sslmode=disable
      - NATS_URL=nats://nats:4222
      - REDIS_ADDR=redis:6379
    depends_on:
      - redis
      - nats
      - postgres

Run:
docker-compose up -d

Troubleshooting

SMTP Error: Verify SMTP_USERNAME and SMTP_PASSWORD:docker run -it --rm pdouble/swaks --to test@example.com --from olzhas200696@gmail.com --server smtp.gmail.com:587 --auth LOGIN --auth-user olzhas200696@gmail.com --auth-password "fdfy ubqq tqps dvcd" -tls


Redis Not Found: Check docker ps for the redis container.
Database Connection: Ensure PostgreSQL is running and the DB connection string is correct.
Service Fails: View logs (e.g., docker logs email) or terminal output.
NATS Issues: Confirm nats container is running and accessible at nats://localhost:4222.

Notes

Production:
Set GIN_MODE=release:export GIN_MODE=release


Use a secure JWT_SECRET and store sensitive data (e.g., SMTP_PASSWORD) in a secrets manager.
Secure Redis with a password and restrict NATS ports.


Windows Paths: Use forward slashes (cmd/email-service/main.go) or double backslashes (cmd\\email-service\\main.go) in PowerShell.

For further assistance, refer to the source code or contact the project maintainer.
