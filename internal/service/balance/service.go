package balance

import (
	"context"
	"time"

	"github.com/gitslim/gophermart/internal/errs"
	"github.com/gitslim/gophermart/internal/models"
	"github.com/gitslim/gophermart/internal/service"
	"github.com/gitslim/gophermart/internal/storage"
)

// BalanceServiceImpl реализует интерфейс service.BalanceService
type BalanceServiceImpl struct {
	userStorage       storage.UserStorage
	withdrawalStorage storage.WithdrawalStorage
}

// NewBalanceService создает новый экземпляр сервиса баланса
func NewBalanceService(userStorage storage.UserStorage, withdrawalStorage storage.WithdrawalStorage) service.BalanceService {
	return &BalanceServiceImpl{
		userStorage:       userStorage,
		withdrawalStorage: withdrawalStorage,
	}
}

// GetBalance возвращает текущий баланс пользователя
func (s *BalanceServiceImpl) GetBalance(ctx context.Context, userID int64) (float64, error) {
	user, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		return 0, errs.NewAppError(errs.ErrInternal, "failed to get user")
	}
	if user == nil {
		return 0, errs.NewAppError(errs.ErrNotFound, "user not found")
	}

	return user.Balance, nil
}

// Withdraw списывает средства с баланса пользователя
func (s *BalanceServiceImpl) Withdraw(ctx context.Context, userID int64, orderNumber string, amount float64) error {
	// Проверяем баланс пользователя
	user, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return errs.NewAppError(errs.ErrUnauthorized, "user not found")
	}

	if user.Balance < amount {
		return errs.NewAppError(errs.ErrPaymentRequired, "insufficient funds")
	}

	// Создаем запись о списании
	withdrawal := &models.Withdrawal{
		UserID:      userID,
		Order:       orderNumber,
		Sum:         amount,
		ProcessedAt: time.Now(),
	}

	if err := s.withdrawalStorage.CreateWithdrawal(ctx, withdrawal); err != nil {
		return errs.NewAppError(errs.ErrInternal, "failed to create withdrawal")
	}

	// Обновляем баланс пользователя
	if err := s.userStorage.UpdateBalance(ctx, userID, -amount); err != nil {
		return errs.NewAppError(errs.ErrInternal, "failed to update balance")
	}

	return nil
}

// GetWithdrawals возвращает историю списаний пользователя
func (s *BalanceServiceImpl) GetWithdrawals(ctx context.Context, userID int64) ([]*models.Withdrawal, error) {
	return s.withdrawalStorage.GetUserWithdrawals(ctx, userID)
}
