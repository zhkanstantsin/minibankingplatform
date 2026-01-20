package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"minibankingplatform/internal/domain"
	"minibankingplatform/pkg/trm"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UsersRepository struct {
	injector *trm.Injector[DBTX]
}

func NewUsersRepository(injector *trm.Injector[DBTX]) *UsersRepository {
	return &UsersRepository{
		injector: injector,
	}
}

func (ur *UsersRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `
		SELECT 
		    id,
		    email,
		    password_hash,
		    created_at,
		    updated_at
		FROM users
		WHERE email = $1
	`

	var (
		id           uuid.UUID
		userEmail    string
		passwordHash string
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := ur.injector.DB(ctx).QueryRow(ctx, query, email).Scan(&id, &userEmail, &passwordHash, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewUserNotFoundError(email)
		}
		return nil, fmt.Errorf("querying user by email: %w", err)
	}

	return domain.NewUserFromDB(domain.UserID(id), userEmail, passwordHash, createdAt, updatedAt), nil
}

func (ur *UsersRepository) GetByID(ctx context.Context, userID domain.UserID) (*domain.User, error) {
	const query = `
		SELECT 
		    id,
		    email,
		    password_hash,
		    created_at,
		    updated_at
		FROM users
		WHERE id = $1
	`

	var (
		id           uuid.UUID
		email        string
		passwordHash string
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := ur.injector.DB(ctx).QueryRow(ctx, query, uuid.UUID(userID)).Scan(&id, &email, &passwordHash, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user with id %v not found", userID)
		}
		return nil, fmt.Errorf("querying user by id: %w", err)
	}

	return domain.NewUserFromDB(domain.UserID(id), email, passwordHash, createdAt, updatedAt), nil
}

func (ur *UsersRepository) Save(ctx context.Context, user *domain.User) error {
	const query = `
		INSERT INTO users (id, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE
		SET 
		    email = EXCLUDED.email,
		    password_hash = EXCLUDED.password_hash,
		    updated_at = EXCLUDED.updated_at
	`

	_, err := ur.injector.DB(ctx).Exec(
		ctx,
		query,
		uuid.UUID(user.ID()),
		user.Email(),
		user.PasswordHash(),
		user.CreatedAt(),
		user.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("upserting user: %w", err)
	}

	return nil
}

func (ur *UsersRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := ur.injector.DB(ctx).QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking user existence: %w", err)
	}

	return exists, nil
}
