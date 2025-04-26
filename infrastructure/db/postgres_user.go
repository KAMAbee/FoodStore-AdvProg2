package db

import (
    "database/sql"
    
    "github.com/google/uuid"
    
    "AdvProg2/domain"
    "AdvProg2/repository"
)

func createUserTableIfNotExist(db *sql.DB) error {
    createUsersTable := `
    CREATE TABLE IF NOT EXISTS users (
        id VARCHAR(36) PRIMARY KEY,
        username VARCHAR(255) NOT NULL UNIQUE,
        password VARCHAR(255) NOT NULL,
        role VARCHAR(50) NOT NULL DEFAULT 'user'
    );
    `

    _, err := db.Exec(createUsersTable)
    return err
}

type PostgresUserRepository struct {
    db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) (*PostgresUserRepository, error) {
    if err := createUserTableIfNotExist(db); err != nil {
        return nil, err
    }

    return &PostgresUserRepository{
        db: db,
    }, nil
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
    
    if user.Role == "" {
        user.Role = "user"
    }

    query := `
        INSERT INTO users (id, username, password, role) 
        VALUES ($1, $2, $3, $4)
    `

    _, err = r.db.Exec(query, user.ID, user.Username, user.Password, user.Role)
    return err
}

func (r *PostgresUserRepository) GetByID(id string) (*domain.User, error) {
    query := `
        SELECT id, username, password, role
        FROM users 
        WHERE id = $1
    `

    var user domain.User
    err := r.db.QueryRow(query, id).Scan(
        &user.ID,
        &user.Username,
        &user.Password,
        &user.Role,
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
        SELECT id, username, password, role
        FROM users 
        WHERE username = $1
    `

    var user domain.User
    err := r.db.QueryRow(query, username).Scan(
        &user.ID,
        &user.Username,
        &user.Password,
        &user.Role,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, repository.ErrUserNotFound
        }
        return nil, err
    }

    return &user, nil
}