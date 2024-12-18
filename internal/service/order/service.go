package order

import (
	"context"
	"net/http"
	"time"

	"github.com/gitslim/gophermart/internal/accrual"
	"github.com/gitslim/gophermart/internal/errs"
	"github.com/gitslim/gophermart/internal/models"
	"github.com/gitslim/gophermart/internal/service"
	"github.com/gitslim/gophermart/internal/storage"
)

// OrderServiceImpl реализует интерфейс service.OrderService
type OrderServiceImpl struct {
	orderStorage  storage.OrderStorage
	userStorage   storage.UserStorage
	accrualClient *accrual.Client
}

// NewOrderService создает новый экземпляр сервиса заказов
func NewOrderService(orderStorage storage.OrderStorage, userStorage storage.UserStorage, accrualClient *accrual.Client) service.OrderService {
	return &OrderServiceImpl{
		orderStorage:  orderStorage,
		userStorage:   userStorage,
		accrualClient: accrualClient,
	}
}

// UploadOrder загружает новый заказ
func (s *OrderServiceImpl) UploadOrder(ctx context.Context, userID int64, orderNumber string) error {
	// Проверяем, существует ли заказ
	existingOrder, err := s.orderStorage.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		return errs.NewAppError(errs.ErrInternal, "failed to get order")
	}
	if existingOrder != nil {
		if existingOrder.UserID == userID {
			return errs.NewAppError(errs.ErrOk, "order already uploaded by the same user")
		}
		return errs.NewAppError(errs.ErrConflict, "order already uploaded by another user")
	}

	// Создаем заказ
	order := &models.Order{
		Number:     orderNumber,
		UserID:     userID,
		Status:     models.OrderStatusNew,
		UploadedAt: time.Now(),
	}

	if err := s.orderStorage.CreateOrder(ctx, order); err != nil {
		return errs.NewAppError(errs.ErrInternal, "failed to create order")
	}

	return nil
}

// GetUserOrders возвращает все заказы пользователя
func (s *OrderServiceImpl) GetUserOrders(ctx context.Context, userID int64) ([]*models.Order, error) {
	return s.orderStorage.GetUserOrders(ctx, userID)
}

// ProcessOrder обрабатывает заказ
func (s *OrderServiceImpl) ProcessOrder(ctx context.Context, orderNumber string) error {
	order, err := s.orderStorage.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		return errs.NewAppError(errs.ErrInternal, "failed to get order")
	}
	if order == nil {
		return errs.NewAppError(errs.ErrNotFound, "order not found")
	}

	// Получаем информацию о начислении от системы расчета баллов
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	accrualResp, statusCode, err := s.accrualClient.GetOrderAccrual(ctx, orderNumber)
	if err != nil {
		return errs.NewAppError(errs.ErrInternal, "failed to get accrual info")
	}
	if statusCode != http.StatusOK {
		return errs.NewAppError(errs.ErrInternal, "failed to get accrual info")
	}

	// Если ответ пустой, значит заказ еще не зарегистрирован в системе начислений
	if accrualResp == nil {
		if err := s.orderStorage.UpdateOrderStatus(ctx, order.ID, models.OrderStatusProcessing, 0); err != nil {
			return errs.NewAppError(errs.ErrInternal, "failed to update order status")
		}
		return nil
	}

	// Обновляем статус и начисление в соответствии с ответом от системы
	if err := s.orderStorage.UpdateOrderStatus(ctx, order.ID, accrualResp.Status, accrualResp.Accrual); err != nil {
		return errs.NewAppError(errs.ErrInternal, "failed to update order")
	}

	// Если заказ обработан и есть начисление, обновляем баланс пользователя
	if accrualResp.Status == models.OrderStatusProcessed && accrualResp.Accrual > 0 {
		if err := s.userStorage.UpdateBalance(ctx, order.UserID, accrualResp.Accrual); err != nil {
			return errs.NewAppError(errs.ErrInternal, "failed to update user balance")
		}
	}

	return nil
}
