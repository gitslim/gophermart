package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gitslim/gophermart/internal/conf"
)

// Response представляет ответ от системы расчета начислений
type Response struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

// Client представляет клиент для взаимодействия с системой расчета начислений
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient создает новый экземпляр клиента системы начислений
func NewClient(config *conf.Config) *Client {
	return &Client{
		baseURL: config.AccrualSystemAddress,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetOrderAccrual получает информацию о начислении баллов за заказ
func (c *Client) GetOrderAccrual(ctx context.Context, orderNumber string) (*Response, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var response Response
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &response, nil
	case http.StatusNoContent:
		return nil, nil
	case http.StatusTooManyRequests:
		// Получаем время ожидания из заголовка
		retryAfter := resp.Header.Get("Retry-After")
		return nil, fmt.Errorf("rate limit exceeded, retry after %s seconds", retryAfter)
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
