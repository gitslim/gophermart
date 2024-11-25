package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"path/filepath"

	"github.com/gitslim/gophermart/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func init() {
	loadQueries()
}

//go:embed sql/*.sql
var sqlFS embed.FS

var (
	CreateUserQuery       string
	GetUserByLoginQuery   string
	GetUserByIDQuery      string
	UpdateBalanceQuery    string
	CreateOrderQuery      string
	GetOrderByNumberQuery string
	GetUserOrdersQuery    string
	UpdateOrderStatus     string
	CreateWithdrawalQuery string
	GetUserWithdrawals    string
	GetOrdersByStatuses   string
)

// PgStorage реализует интерфейс storage.Storage для PostgreSQL
type PgStorage struct {
	db *pgxpool.Pool
}

// loadQueries загружает SQL-запросы из файлов и присваивает их переменным.
func loadQueries() {
	queries := map[string]*string{
		"create_user.sql":          &CreateUserQuery,
		"get_user_by_login.sql":    &GetUserByLoginQuery,
		"get_user_by_id.sql":       &GetUserByIDQuery,
		"update_balance.sql":       &UpdateBalanceQuery,
		"create_order.sql":         &CreateOrderQuery,
		"get_order_by_number.sql":  &GetOrderByNumberQuery,
		"get_user_orders.sql":      &GetUserOrdersQuery,
		"update_order_status.sql":  &UpdateOrderStatus,
		"create_withdrawal.sql":    &CreateWithdrawalQuery,
		"get_user_withdrawals.sql": &GetUserWithdrawals,
		"get_orders_by_statuses.sql": &GetOrdersByStatuses,
	}

	for file, qPtr := range queries {
		data, err := sqlFS.ReadFile(filepath.Join("sql", file))
		if err != nil {
			log.Fatalf("Ошибка загрузки SQL-запроса из файла %s: %v", file, err)
		}
		*qPtr = string(data)
	}
}

// NewPgStorage создает новый экземпляр хранилища PostgreSQL
func NewPgStorage(pool *pgxpool.Pool) *PgStorage {
	return &PgStorage{
		db: pool,
	}
}

// CreateUser создает нового пользователя
func (s *PgStorage) CreateUser(ctx context.Context, user *models.User) error {
	row := s.db.QueryRow(ctx, CreateUserQuery,
		user.Login,
		user.PasswordHash,
		user.Balance,
		user.CreatedAt,
	)

	return row.Scan(&user.ID)
}

// GetUserByLogin возвращает пользователя по логину
func (s *PgStorage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
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
func (s *PgStorage) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
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
func (s *PgStorage) UpdateBalance(ctx context.Context, userID int64, delta float64) error {

	_, err := s.db.Exec(ctx, UpdateBalanceQuery, userID, delta)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	return nil
}

// CreateOrder создает новый заказ
func (s *PgStorage) CreateOrder(ctx context.Context, order *models.Order) error {

	_, err := s.db.Exec(ctx, CreateOrderQuery,
		order.Number,
		order.UserID,
		order.Status,
		order.Accrual,
		order.UploadedAt,
		order.ProcessedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

// GetOrderByNumber возвращает заказ по номеру
func (s *PgStorage) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	order := &models.Order{}

	err := s.db.QueryRow(ctx, GetOrderByNumberQuery, number).Scan(
		&order.ID,
		&order.Number,
		&order.UserID,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
		&order.ProcessedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order by number: %w", err)
	}

	return order, nil
}

// GetUserOrders возвращает все заказы пользователя
func (s *PgStorage) GetUserOrders(ctx context.Context, userID int64) ([]*models.Order, error) {

	rows, err := s.db.Query(ctx, GetUserOrdersQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		err := rows.Scan(
			&order.ID,
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
			&order.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}

	return orders, nil
}

// UpdateOrderStatus обновляет статус заказа
func (s *PgStorage) UpdateOrderStatus(ctx context.Context, orderID int64, status string, accrual float64) error {

	_, err := s.db.Exec(ctx, UpdateOrderStatus, orderID, status, accrual)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

// CreateWithdrawal создает новую операцию списания
func (s *PgStorage) CreateWithdrawal(ctx context.Context, withdrawal *models.Withdrawal) error {

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
func (s *PgStorage) GetUserWithdrawals(ctx context.Context, userID int64) ([]*models.Withdrawal, error) {

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

// GetOrdersByStatuses возвращает заказы с указанными статусами
func (s *PgStorage) GetOrdersByStatuses(ctx context.Context, statuses []string) ([]*models.Order, error) {
	rows, err := s.db.Query(ctx, GetOrdersByStatuses, statuses)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by statuses: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		err := rows.Scan(
			&order.ID,
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
			&order.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}

	return orders, nil
}
