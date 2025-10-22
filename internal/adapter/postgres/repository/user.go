package repository

import (
	"context"
	"database/sql"
	"fmt"
	"webike_services/webike_User-microservice_Nikita/internal/core/domain"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db,
	}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `INSERT INTO users (name, date_of_birth, email, password, role)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query, user.Name, user.DateOfBirth, user.Email, user.Password, user.Role).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				return nil, fmt.Errorf("email already exists")
			case "23502":
				return nil, fmt.Errorf("required field is missing")
			default:
				return nil, err
			}
		}
		return nil, err
	}
	return user, nil
}

func (r *PostgresUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, name, date_of_birth, email, password, created_at, updated_at, role
              FROM users WHERE id = $1`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.DateOfBirth,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Role,
	)

	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *PostgresUserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("User not found")
	}

	return nil
}

func (r *PostgresUserRepository) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `UPDATE users
        SET 
        name = COALESCE(NULLIF($1, ''), name),
        date_of_birth = COALESCE(NULLIF($2, ''), date_of_birth),
        email = COALESCE(NULLIF($3, ''), email),
        password = COALESCE(NULLIF($4, ''), password),
        updated_at = CURRENT_TIMESTAMP
        WHERE id = $5
        RETURNING id, name, date_of_birth, email, password, created_at, updated_at, role`

	result := &domain.User{}
	err := r.db.QueryRowContext(ctx, query,
		user.Name, user.DateOfBirth, user.Email, user.Password, user.ID).Scan(
		&result.ID,
		&result.Name,
		&result.DateOfBirth,
		&result.Email,
		&result.Password,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.Role,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, fmt.Errorf("email already exists")
		}
		return nil, fmt.Errorf("Error updating user: %w", err)
	}
	return result, nil
}
func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, name, date_of_birth, email, password, created_at, updated_at, role
              FROM users WHERE email = $1`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.DateOfBirth,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Role,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}
