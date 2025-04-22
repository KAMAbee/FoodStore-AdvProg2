package db

import (
    "database/sql"
    "errors"
    "fmt"
    "strings"
    
    "AdvProg2/domain"
    _ "github.com/lib/pq"
)

type PostgresProductRepository struct {
    db *sql.DB
}

func NewPostgresProductRepository(db *sql.DB) *PostgresProductRepository {
    return &PostgresProductRepository{
        db: db,
    }
}

func (r *PostgresProductRepository) Create(product *domain.Product) error {
    query := `INSERT INTO products (id, name, price, stock) VALUES ($1, $2, $3, $4)`
    
    _, err := r.db.Exec(query, product.ID, product.Name, product.Price, product.Stock)
    return err
}

func (r *PostgresProductRepository) GetByID(id string) (*domain.Product, error) {
    query := `SELECT id, name, price, stock FROM products WHERE id = $1`
    
    var product domain.Product
    err := r.db.QueryRow(query, id).Scan(&product.ID, &product.Name, &product.Price, &product.Stock)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.New("product not found")
        }
        return nil, err
    }
    
    return &product, nil
}

func (r *PostgresProductRepository) Update(product *domain.Product) error {
    query := `UPDATE products SET name = $2, price = $3, stock = $4 WHERE id = $1`
    
    res, err := r.db.Exec(query, product.ID, product.Name, product.Price, product.Stock)
    if err != nil {
        return err
    }
    
    rowsAffected, err := res.RowsAffected()
    if err != nil {
        return err
    }
    
    if rowsAffected == 0 {
        return errors.New("product not found")
    }
    
    return nil
}

func (r *PostgresProductRepository) Delete(id string) error {
    query := `DELETE FROM products WHERE id = $1`
    
    res, err := r.db.Exec(query, id)
    if err != nil {
        return err
    }
    
    rowsAffected, err := res.RowsAffected()
    if err != nil {
        return err
    }
    
    if rowsAffected == 0 {
        return errors.New("product not found")
    }
    
    return nil
}

func (r *PostgresProductRepository) List(page, limit int32) ([]*domain.Product, int32, error) {
    offset := (page - 1) * limit
    
    countQuery := `SELECT COUNT(*) FROM products`
    var total int32
    err := r.db.QueryRow(countQuery).Scan(&total)
    if err != nil {
        return nil, 0, err
    }
    
    query := `SELECT id, name, price, stock FROM products ORDER BY name LIMIT $1 OFFSET $2`
    
    rows, err := r.db.Query(query, limit, offset)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()
    
    var products []*domain.Product
    
    for rows.Next() {
        var product domain.Product
        if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.Stock); err != nil {
            return nil, 0, err
        }
        products = append(products, &product)
    }
    
    if err = rows.Err(); err != nil {
        return nil, 0, err
    }
    
    return products, total, nil
}

func (r *PostgresProductRepository) SearchByName(name string, page, limit int32) ([]*domain.Product, int32, error) {
    offset := (page - 1) * limit
    
    searchPattern := "%" + strings.ToLower(name) + "%"
    
    countQuery := `SELECT COUNT(*) FROM products WHERE LOWER(name) LIKE $1`
    var total int32
    err := r.db.QueryRow(countQuery, searchPattern).Scan(&total)
    if err != nil {
        return nil, 0, err
    }
    
    query := `SELECT id, name, price, stock FROM products WHERE LOWER(name) LIKE $1 ORDER BY name LIMIT $2 OFFSET $3`
    
    rows, err := r.db.Query(query, searchPattern, limit, offset)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()
    
    var products []*domain.Product
    
    for rows.Next() {
        var product domain.Product
        if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.Stock); err != nil {
            return nil, 0, err
        }
        products = append(products, &product)
    }
    
    if err = rows.Err(); err != nil {
        return nil, 0, err
    }
    
    return products, total, nil
}

func (r *PostgresProductRepository) SearchByPriceRange(minPrice, maxPrice float64, page, limit int32) ([]*domain.Product, int32, error) {
    offset := (page - 1) * limit
    
    var countQuery string
    var countArgs []interface{}
    
    if minPrice > 0 && maxPrice > 0 {
        countQuery = `SELECT COUNT(*) FROM products WHERE price >= $1 AND price <= $2`
        countArgs = []interface{}{minPrice, maxPrice}
    } else if minPrice > 0 {
        countQuery = `SELECT COUNT(*) FROM products WHERE price >= $1`
        countArgs = []interface{}{minPrice}
    } else if maxPrice > 0 {
        countQuery = `SELECT COUNT(*) FROM products WHERE price <= $1`
        countArgs = []interface{}{maxPrice}
    } else {
        countQuery = `SELECT COUNT(*) FROM products`
    }
    
    var total int32
    err := r.db.QueryRow(countQuery, countArgs...).Scan(&total)
    if err != nil {
        return nil, 0, err
    }
    
    var query string
    var queryArgs []interface{}
    
    if minPrice > 0 && maxPrice > 0 {
        query = `SELECT id, name, price, stock FROM products WHERE price >= $1 AND price <= $2 ORDER BY name LIMIT $3 OFFSET $4`
        queryArgs = []interface{}{minPrice, maxPrice, limit, offset}
    } else if minPrice > 0 {
        query = `SELECT id, name, price, stock FROM products WHERE price >= $1 ORDER BY name LIMIT $2 OFFSET $3`
        queryArgs = []interface{}{minPrice, limit, offset}
    } else if maxPrice > 0 {
        query = `SELECT id, name, price, stock FROM products WHERE price <= $1 ORDER BY name LIMIT $2 OFFSET $3`
        queryArgs = []interface{}{maxPrice, limit, offset}
    } else {
        query = `SELECT id, name, price, stock FROM products ORDER BY name LIMIT $1 OFFSET $2`
        queryArgs = []interface{}{limit, offset}
    }
    
    rows, err := r.db.Query(query, queryArgs...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()
    
    var products []*domain.Product
    
    for rows.Next() {
        var product domain.Product
        if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.Stock); err != nil {
            return nil, 0, err
        }
        products = append(products, &product)
    }
    
    if err = rows.Err(); err != nil {
        return nil, 0, err
    }
    
    return products, total, nil
}

func (r *PostgresProductRepository) SearchByFilters(name string, minPrice, maxPrice float64, page, limit int32) ([]*domain.Product, int32, error) {
    offset := (page - 1) * limit
    
    searchPattern := "%" + strings.ToLower(name) + "%"
    
    var conditions []string
    var args []interface{}
    var argIndex int = 1
    
    if name != "" {
        conditions = append(conditions, fmt.Sprintf("LOWER(name) LIKE $%d", argIndex))
        args = append(args, searchPattern)
        argIndex++
    }
    
    if minPrice > 0 {
        conditions = append(conditions, fmt.Sprintf("price >= $%d", argIndex))
        args = append(args, minPrice)
        argIndex++
    }
    
    if maxPrice > 0 {
        conditions = append(conditions, fmt.Sprintf("price <= $%d", argIndex))
        args = append(args, maxPrice)
        argIndex++
    }
    
    whereClause := ""
    if len(conditions) > 0 {
        whereClause = "WHERE " + strings.Join(conditions, " AND ")
    }
    
    countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products %s", whereClause)
    var total int32
    err := r.db.QueryRow(countQuery, args...).Scan(&total)
    if err != nil {
        return nil, 0, err
    }
    
    query := fmt.Sprintf("SELECT id, name, price, stock FROM products %s ORDER BY name LIMIT $%d OFFSET $%d", 
                         whereClause, argIndex, argIndex+1)
    args = append(args, limit, offset)
    
    rows, err := r.db.Query(query, args...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()
    
    var products []*domain.Product
    
    for rows.Next() {
        var product domain.Product
        if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.Stock); err != nil {
            return nil, 0, err
        }
        products = append(products, &product)
    }
    
    if err = rows.Err(); err != nil {
        return nil, 0, err
    }
    
    return products, total, nil
}