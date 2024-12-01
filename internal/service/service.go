package service

import (
	"context"

	"github.com/gitslim/gophermart/internal/models"
)

// UserService определяет интерфейс для работы с пользователями
type UserService interface {
	Register(ctx context.Context, login, password string) (*models.User, error)
	Login(ctx context.Context, login, password string) (*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
}

// OrderService определяет интерфейс для работы с заказами
type OrderService interface {
	UploadOrder(ctx context.Context, userID int64, orderNumber string) error
	GetUserOrders(ctx context.Context, userID int64) ([]*models.Order, error)
	ProcessOrder(ctx context.Context, orderNumber string) error
}

// BalanceService определяет интерфейс для работы с балансом
type BalanceService interface {
	GetBalance(ctx context.Context, userID int64) (float64, error)
	Withdraw(ctx context.Context, userID int64, orderNumber string, amount float64) error
	GetWithdrawals(ctx context.Context, userID int64) ([]*models.Withdrawal, error)
}
