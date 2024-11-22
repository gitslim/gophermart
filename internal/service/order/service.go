package order

import (
	"context"
	"fmt"
	"time"

	"github.com/gitslim/gophermart/internal/models"
	"github.com/gitslim/gophermart/internal/service"
	"github.com/gitslim/gophermart/internal/storage"
)

// OrderServiceImpl реализует интерфейс service.OrderService
type OrderServiceImpl struct {
	storage storage.Storage
}

// NewOrderService создает новый экземпляр сервиса заказов
func NewOrderService(storage storage.Storage) service.OrderService {
	return &OrderServiceImpl{
		storage: storage,
	}
}

// UploadOrder загружает новый заказ
func (s *OrderServiceImpl) UploadOrder(ctx context.Context, userID int64, orderNumber string) error {
	// Проверяем, существует ли заказ
	existingOrder, err := s.storage.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		return fmt.Errorf("failed to check existing order: %w", err)
	}
	if existingOrder != nil {
		if existingOrder.UserID == userID {
			return fmt.Errorf("order already uploaded by this user")
		}
		return fmt.Errorf("order already uploaded by another user")
	}

	// Создаем заказ
	order := &models.Order{
		Number:     orderNumber,
		UserID:     userID,
		Status:     models.OrderStatusNew,
		UploadedAt: time.Now(),
	}

	if err := s.storage.CreateOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

// GetUserOrders возвращает все заказы пользователя
func (s *OrderServiceImpl) GetUserOrders(ctx context.Context, userID int64) ([]*models.Order, error) {
	return s.storage.GetUserOrders(ctx, userID)
}

// ProcessOrder обрабатывает заказ
func (s *OrderServiceImpl) ProcessOrder(ctx context.Context, orderNumber string) error {
	order, err := s.storage.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return fmt.Errorf("order not found")
	}

	// TODO: Здесь должна быть логика взаимодействия с системой начисления баллов
	// Пока просто обновляем статус
	if err := s.storage.UpdateOrderStatus(ctx, order.ID, models.OrderStatusProcessed, 0); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}
