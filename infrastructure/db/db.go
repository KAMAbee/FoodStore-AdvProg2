package db

import (
    "database/sql"
    "fmt"
    "log"
    "os"

    _ "github.com/lib/pq"
)

func NewPostgresConnection() (*sql.DB, error) {
    dbURL := os.Getenv("DB")
    if dbURL == "" {
        dbURL = "postgresql://postgres:admin@localhost:5432/FoodStore?sslmode=disable"
    }

    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        return nil, fmt.Errorf("error connecting to database: %w", err)
    }

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("error pinging database: %w", err)
    }

    if err := createTablesIfNotExist(db); err != nil {
        return nil, fmt.Errorf("error creating tables: %w", err)
    }

    log.Println("Connected to Postgres database")
    return db, nil
}

func createTablesIfNotExist(db *sql.DB) error {
    createProductsTable := `
    CREATE TABLE IF NOT EXISTS products (
        ID VARCHAR(36) PRIMARY KEY,
        Name VARCHAR(255) NOT NULL,
        Price DECIMAL(10, 2) NOT NULL,
        Stock INT NOT NULL
    );
    `

    _, err := db.Exec(createProductsTable)
    if err != nil {
        return err
    }

    return nil
}