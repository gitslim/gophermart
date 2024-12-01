package web

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gitslim/gophermart/internal/conf"
	"github.com/gitslim/gophermart/internal/logging"
	"go.uber.org/fx"
)

// RegisterServerHooks регистрирует хуки для запуска и остановки HTTP сервера
func RegisterServerHooks(lc fx.Lifecycle, cfg *conf.Config, log logging.Logger, router *gin.Engine) {
	srv := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: router,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Infof("Starting HTTP server on %v", srv.Addr)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatalf("HTTP server failed: %s", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping HTTP server")
			return srv.Shutdown(ctx)
		},
	})
}
