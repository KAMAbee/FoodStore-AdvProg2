package db

import (
    "database/sql"
    "errors"
    "fmt"
    "time"
    
    "github.com/google/uuid"
    "AdvProg2/domain"
)

func createOrderTablesIfNotExist(db *sql.DB) error {
    createOrdersTable := `
    CREATE TABLE IF NOT EXISTS orders (
        id VARCHAR(36) PRIMARY KEY,
        user_id VARCHAR(255) NOT NULL,
        status VARCHAR(50) NOT NULL DEFAULT 'pending',
        total_price DECIMAL(10, 2) NOT NULL DEFAULT 0,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );
    `

    createOrderItemsTable := `
    CREATE TABLE IF NOT EXISTS order_items (
        id VARCHAR(36) PRIMARY KEY,
        order_id VARCHAR(36) NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
        product_id VARCHAR(36) NOT NULL REFERENCES products(id),
        quantity INT NOT NULL,
        price DECIMAL(10, 2) NOT NULL,
        CONSTRAINT unique_order_product UNIQUE (order_id, product_id)
    );
    `

    _, err := db.Exec(createOrdersTable)
    if err != nil {
        return err
    }

    _, err = db.Exec(createOrderItemsTable)
    if err != nil {
        return err
    }

    return nil
}

type PostgresOrderRepository struct {
    db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
    err := createOrderTablesIfNotExist(db)
    if err != nil {
        panic(fmt.Sprintf("Failed to create order tables: %v", err))
    }
    
    return &PostgresOrderRepository{
        db: db,
    }
}

func (r *PostgresOrderRepository) Create(order *domain.Order) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer func() {
        if err != nil {
            tx.Rollback()
            return
        }
        err = tx.Commit()
    }()

    query := `
        INSERT INTO orders (id, user_id, status, total_price, created_at) 
        VALUES ($1, $2, $3, $4, $5)
    `
    
    if order.ID == "" {
        order.ID = uuid.New().String()
    }
    
    if order.CreatedAt.IsZero() {
        order.CreatedAt = time.Now()
    }
    
    _, err = tx.Exec(query, order.ID, order.UserID, order.Status, order.TotalPrice, order.CreatedAt)
    if err != nil {
        return err
    }
    
    for _, item := range order.Items {
        if item.ID == "" {
            item.ID = uuid.New().String()
        }
        
        item.OrderID = order.ID
        
        query := `
            INSERT INTO order_items (id, order_id, product_id, quantity, price) 
            VALUES ($1, $2, $3, $4, $5)
        `
        _, err = tx.Exec(query, item.ID, item.OrderID, item.ProductID, item.Quantity, item.Price)
        if err != nil {
            return err
        }
    }
    
    return nil
}

func (r *PostgresOrderRepository) GetByID(id string) (*domain.Order, error) {
    orderQuery := `
        SELECT id, user_id, status, total_price, created_at 
        FROM orders 
        WHERE id = $1
    `
    
    var order domain.Order
    err := r.db.QueryRow(orderQuery, id).Scan(
        &order.ID, 
        &order.UserID, 
        &order.Status, 
        &order.TotalPrice, 
        &order.CreatedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.New("order not found")
        }
        return nil, err
    }
    
    itemsQuery := `
        SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price,
               p.id, p.name, p.price, p.stock
        FROM order_items oi
        LEFT JOIN products p ON oi.product_id = p.id
        WHERE oi.order_id = $1
    `
    
    rows, err := r.db.Query(itemsQuery, id)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    for rows.Next() {
        var item domain.OrderItem
        var product domain.Product
        
        err := rows.Scan(
            &item.ID,
            &item.OrderID,
            &item.ProductID,
            &item.Quantity,
            &item.Price,
            &product.ID,
            &product.Name,
            &product.Price,
            &product.Stock,
        )
        
        if err != nil {
            return nil, err
        }
        
        item.Product = &product
        order.Items = append(order.Items, &item)
    }
    
    if err = rows.Err(); err != nil {
        return nil, err
    }
    
    return &order, nil
}

func (r *PostgresOrderRepository) GetByUserID(userID string, page, limit int32) ([]*domain.Order, int32, error) {
    offset := (page - 1) * limit
    
    countQuery := "SELECT COUNT(*) FROM orders WHERE user_id = $1"
    var total int32
    err := r.db.QueryRow(countQuery, userID).Scan(&total)
    if err != nil {
        return nil, 0, err
    }
    
    ordersQuery := `
        SELECT id, user_id, status, total_price, created_at 
        FROM orders 
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `
    
    rows, err := r.db.Query(ordersQuery, userID, limit, offset)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()
    
    var orders []*domain.Order
    
    for rows.Next() {
        var order domain.Order
        err := rows.Scan(
            &order.ID,
            &order.UserID,
            &order.Status,
            &order.TotalPrice,
            &order.CreatedAt,
        )
        
        if err != nil {
            return nil, 0, err
        }
        
        orderDetails, err := r.GetByID(order.ID)
        if err != nil {
            return nil, 0, err
        }
        
        order.Items = orderDetails.Items
        orders = append(orders, &order)
    }
    
    if err = rows.Err(); err != nil {
        return nil, 0, err
    }
    
    return orders, total, nil
}

func (r *PostgresOrderRepository) UpdateStatus(id, status string) error {
    query := "UPDATE orders SET status = $1 WHERE id = $2"
    
    res, err := r.db.Exec(query, status, id)
    if err != nil {
        return err
    }
    
    rowsAffected, err := res.RowsAffected()
    if err != nil {
        return err
    }
    
    if rowsAffected == 0 {
        return errors.New("order not found")
    }
    
    return nil
}

func (r *PostgresOrderRepository) Delete(id string) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer func() {
        if err != nil {
            tx.Rollback()
            return
        }
        err = tx.Commit()
    }()
    
    query := "DELETE FROM orders WHERE id = $1"
    
    res, err := tx.Exec(query, id)
    if err != nil {
        return err
    }
    
    rowsAffected, err := res.RowsAffected()
    if err != nil {
        return err
    }
    
    if rowsAffected == 0 {
        return errors.New("order not found")
    }
    
    return nil
}