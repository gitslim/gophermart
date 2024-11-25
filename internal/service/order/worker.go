package order

import (
	"context"
	"time"

	"github.com/gitslim/gophermart/internal/logging"
	"github.com/gitslim/gophermart/internal/models"
	"github.com/gitslim/gophermart/internal/service"
	"github.com/gitslim/gophermart/internal/storage"
	"go.uber.org/fx"
)

// Worker представляет фоновый обработчик заказов
type Worker struct {
	service service.OrderService
	storage storage.Storage
	log     logging.Logger
}

// NewWorker создает новый экземпляр фонового обработчика заказов
func NewWorker(service service.OrderService, storage storage.Storage, logger logging.Logger) *Worker {
	return &Worker{
		service: service,
		storage: storage,
		log:     logger,
	}
}

// Start запускает фоновую обработку заказов
func (w *Worker) Start(ctx context.Context) error {
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
func (w *Worker) processOrders(ctx context.Context) error {
	// Получаем все заказы в статусе NEW или PROCESSING
	orders, err := w.storage.GetOrdersByStatuses(ctx, []string{
		models.OrderStatusNew,
		models.OrderStatusProcessing,
	})
	if err != nil {
		return err
	}

	// Обрабатываем каждый заказ
	for _, order := range orders {
		if err := w.service.ProcessOrder(ctx, order.Number); err != nil {
			w.log.Errorf("Failed to process order %s: %v", order.Number, err)
			continue
		}
	}

	return nil
}

// RegisterWorkerHooks регистрирует хуки для запуска и остановки воркера
func RegisterWorkerHooks(lc fx.Lifecycle, worker *Worker) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go worker.Start(ctx)
			return nil
		},
	})
}
