package db

import (
    "database/sql"
    "fmt"

    "github.com/google/uuid"
    _ "github.com/lib/pq"
    
    "AdvProg2/domain"
    "AdvProg2/repository"
)

func createUserTableIfNotExist(db *sql.DB) error {
    createUsersTable := `
    CREATE TABLE IF NOT EXISTS users (
        id VARCHAR(36) PRIMARY KEY,
        username VARCHAR(255) NOT NULL UNIQUE,
        password VARCHAR(255) NOT NULL
    );
    `

    _, err := db.Exec(createUsersTable)
    return err
}

type PostgresUserRepository struct {
    db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) repository.UserRepository {
    err := createUserTableIfNotExist(db)
    if err != nil {
        panic(fmt.Sprintf("Failed to create users table: %v", err))
    }

    return &PostgresUserRepository{
        db: db,
    }
}

func (r *PostgresUserRepository) Create(user *domain.User) error {
    var count int
    err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", user.Username).Scan(&count)
    if err != nil {
        return err
    }

    if count > 0 {
        return repository.ErrUsernameAlreadyExists
    }

    if user.ID == "" {
        user.ID = uuid.New().String()
    }

    query := `
        INSERT INTO users (id, username, password) 
        VALUES ($1, $2, $3)
    `

    _, err = r.db.Exec(query, user.ID, user.Username, user.Password)
    return err
}

func (r *PostgresUserRepository) GetByID(id string) (*domain.User, error) {
    query := `
        SELECT id, username, password 
        FROM users 
        WHERE id = $1
    `

    var user domain.User
    err := r.db.QueryRow(query, id).Scan(
        &user.ID,
        &user.Username,
        &user.Password,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, repository.ErrUserNotFound
        }
        return nil, err
    }

    return &user, nil
}

func (r *PostgresUserRepository) GetByUsername(username string) (*domain.User, error) {
    query := `
        SELECT id, username, password 
        FROM users 
        WHERE username = $1
    `

    var user domain.User
    err := r.db.QueryRow(query, username).Scan(
        &user.ID,
        &user.Username,
        &user.Password,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, repository.ErrUserNotFound
        }
        return nil, err
    }

    return &user, nil
}