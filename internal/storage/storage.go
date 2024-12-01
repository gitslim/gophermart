package storage

import (
	"context"

	"github.com/gitslim/gophermart/internal/models"
)

// UserStorage определяет интерфейс для работы с пользователями
type UserStorage interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	UpdateBalance(ctx context.Context, userID int64, delta float64) error
}

// OrderStorage определяет интерфейс для работы с заказами
type OrderStorage interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	GetOrderByNumber(ctx context.Context, number string) (*models.Order, error)
	GetUserOrders(ctx context.Context, userID int64) ([]*models.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, status string, accrual float64) error
	GetOrdersByStatuses(ctx context.Context, statuses []string) ([]*models.Order, error)
}

// WithdrawalStorage определяет интерфейс для работы со списаниями
type WithdrawalStorage interface {
	CreateWithdrawal(ctx context.Context, withdrawal *models.Withdrawal) error
	GetUserWithdrawals(ctx context.Context, userID int64) ([]*models.Withdrawal, error)
}
