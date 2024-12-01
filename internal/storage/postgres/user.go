package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gitslim/gophermart/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	CreateUserQuery     string
	GetUserByLoginQuery string
	GetUserByIDQuery    string
	UpdateBalanceQuery  string
)

func init() {
	queries := map[string]*string{
		"create_user.sql":       &CreateUserQuery,
		"get_user_by_login.sql": &GetUserByLoginQuery,
		"get_user_by_id.sql":    &GetUserByIDQuery,
		"update_balance.sql":    &UpdateBalanceQuery,
	}
	loadQueries(queries)
}

// PgUserStorage представляет хранилище пользователей PostgreSQL
type PgUserStorage struct {
	db *pgxpool.Pool
}

// NewPgUserStorage создает новый экземпляр хранилища PostgreSQL
func NewPgUserStorage(pool *pgxpool.Pool) *PgUserStorage {
	return &PgUserStorage{
		db: pool,
	}
}

// CreateUser создает нового пользователя
func (s *PgUserStorage) CreateUser(ctx context.Context, user *models.User) error {
	row := s.db.QueryRow(ctx, CreateUserQuery,
		user.Login,
		user.PasswordHash,
		user.Balance,
		user.CreatedAt,
	)

	return row.Scan(&user.ID)
}

// GetUserByLogin возвращает пользователя по логину
func (s *PgUserStorage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	user := &models.User{}

	err := s.db.QueryRow(ctx, GetUserByLoginQuery, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.Balance,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by login: %w", err)
	}

	return user, nil
}

// GetUserByID возвращает пользователя по ID
func (s *PgUserStorage) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	user := &models.User{}

	err := s.db.QueryRow(ctx, GetUserByIDQuery, id).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.Balance,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}

// UpdateBalance обновляет баланс пользователя
func (s *PgUserStorage) UpdateBalance(ctx context.Context, userID int64, delta float64) error {

	_, err := s.db.Exec(ctx, UpdateBalanceQuery, userID, delta)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	return nil
}
