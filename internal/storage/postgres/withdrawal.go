package postgres

import (
	"context"
	"fmt"

	"github.com/gitslim/gophermart/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	CreateWithdrawalQuery string
	GetUserWithdrawals    string
)

func init() {
	queries := map[string]*string{
		"create_withdrawal.sql":    &CreateWithdrawalQuery,
		"get_user_withdrawals.sql": &GetUserWithdrawals,
	}

	loadQueries(queries)
}

// PgWithdrawalStorage представляет хранилище операций списания
type PgWithdrawalStorage struct {
	db *pgxpool.Pool
}

// NewPgWithdrawalStorage создает новый экземпляр хранилища PostgreSQL
func NewPgWithdrawalStorage(pool *pgxpool.Pool) *PgWithdrawalStorage {
	return &PgWithdrawalStorage{
		db: pool,
	}
}

// CreateWithdrawal создает новую операцию списания
func (s *PgWithdrawalStorage) CreateWithdrawal(ctx context.Context, withdrawal *models.Withdrawal) error {

	_, err := s.db.Exec(ctx, CreateWithdrawalQuery,
		withdrawal.UserID,
		withdrawal.Order,
		withdrawal.Sum,
		withdrawal.ProcessedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create withdrawal: %w", err)
	}

	return nil
}

// GetUserWithdrawals возвращает все операции списания пользователя
func (s *PgWithdrawalStorage) GetUserWithdrawals(ctx context.Context, userID int64) ([]*models.Withdrawal, error) {

	rows, err := s.db.Query(ctx, GetUserWithdrawals, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user withdrawals: %w", err)
	}
	defer rows.Close()

	var withdrawals []*models.Withdrawal
	for rows.Next() {
		withdrawal := &models.Withdrawal{}
		err := rows.Scan(
			&withdrawal.ID,
			&withdrawal.UserID,
			&withdrawal.Order,
			&withdrawal.Sum,
			&withdrawal.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating withdrawals: %w", err)
	}

	return withdrawals, nil
}
