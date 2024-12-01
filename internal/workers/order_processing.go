package workers

import (
	"context"
	"time"

	"github.com/gitslim/gophermart/internal/logging"
	"github.com/gitslim/gophermart/internal/models"
	"github.com/gitslim/gophermart/internal/service"
	"github.com/gitslim/gophermart/internal/storage"
	"go.uber.org/fx"
)

// OrderProcessingWorker представляет фоновый обработчик заказов
type OrderProcessingWorker struct {
	orderService service.OrderService
	orderStorage storage.OrderStorage
	log     logging.Logger
}

// NewOrderProcessingWorker создает новый экземпляр фонового обработчика заказов
func NewOrderProcessingWorker(orderService service.OrderService, sorderStorage storage.OrderStorage, log logging.Logger) *OrderProcessingWorker {
	return &OrderProcessingWorker{
		orderService: orderService,
		orderStorage: sorderStorage,
		log:     log,
	}
}

// Start запускает фоновую обработку заказов
func (w *OrderProcessingWorker) Start(ctx context.Context) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.processOrders(ctx); err != nil {
				w.log.Errorf("Failed to process orders: %v", err)
			}
		}
	}
}

// processOrders обрабатывает все необработанные заказы
func (w *OrderProcessingWorker) processOrders(ctx context.Context) error {
	// Получаем все заказы в статусе NEW или PROCESSING
	orders, err := w.orderStorage.GetOrdersByStatuses(ctx, []string{
		models.OrderStatusNew,
		models.OrderStatusProcessing,
	})
	if err != nil {
		return err
	}

	// Обрабатываем каждый заказ
	for _, order := range orders {
		if err := w.orderService.ProcessOrder(ctx, order.Number); err != nil {
			w.log.Errorf("Failed to process order %s: %v", order.Number, err)
			continue
		}
	}

	return nil
}

// RegisterOrderProcessingWorkerHooks регистрирует хуки для запуска и остановки воркера
func RegisterOrderProcessingWorkerHooks(lc fx.Lifecycle, worker *OrderProcessingWorker) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go worker.Start(ctx)
			return nil
		},
	})
}
