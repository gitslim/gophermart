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
	CreateOrderQuery      string
	GetOrderByNumberQuery string
	GetUserOrdersQuery    string
	UpdateOrderStatus     string
	GetOrdersByStatuses   string
)

func init() {
	queries := map[string]*string{
		"create_order.sql":           &CreateOrderQuery,
		"get_order_by_number.sql":    &GetOrderByNumberQuery,
		"get_user_orders.sql":        &GetUserOrdersQuery,
		"update_order_status.sql":    &UpdateOrderStatus,
		"get_orders_by_statuses.sql": &GetOrdersByStatuses,
	}

	loadQueries(queries)
}

// PgOrderStorage представляет хранилище заказов в PostgreSQL
type PgOrderStorage struct {
	db *pgxpool.Pool
}

// NewPgOrderStorage создает новый экземпляр хранилища PostgreSQL
func NewPgOrderStorage(pool *pgxpool.Pool) *PgOrderStorage {
	return &PgOrderStorage{
		db: pool,
	}
}

// CreateOrder создает новый заказ
func (s *PgOrderStorage) CreateOrder(ctx context.Context, order *models.Order) error {

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
func (s *PgOrderStorage) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
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
func (s *PgOrderStorage) GetUserOrders(ctx context.Context, userID int64) ([]*models.Order, error) {

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
func (s *PgOrderStorage) UpdateOrderStatus(ctx context.Context, orderID int64, status string, accrual float64) error {

	_, err := s.db.Exec(ctx, UpdateOrderStatus, orderID, status, accrual)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

// GetOrdersByStatuses возвращает заказы с указанными статусами
func (s *PgOrderStorage) GetOrdersByStatuses(ctx context.Context, statuses []string) ([]*models.Order, error) {
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
