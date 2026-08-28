package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/andrey1af/shop-api/api-service/internal/config"
	"github.com/andrey1af/shop-api/api-service/internal/database"
	"github.com/andrey1af/shop-api/api-service/internal/handlers"
	"github.com/andrey1af/shop-api/api-service/internal/repositories"
	"github.com/andrey1af/shop-api/api-service/internal/services"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	applicationContext, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	startupContext, cancelStartup := context.WithTimeout(
		applicationContext,
		cfg.DatabasePool.ConnectTimeout,
	)
	pool, err := database.NewPostgresPool(startupContext, cfg.DatabaseURL, cfg.DatabasePool)
	cancelStartup()
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	supplierRepository := repositories.NewSupplierRepository(pool)
	supplierService := services.NewSupplierService(supplierRepository)
	clientRepository := repositories.NewClientRepository(pool)
	clientService := services.NewClientService(clientRepository)
	productRepository := repositories.NewProductRepository(pool)
	productService := services.NewProductService(productRepository)
	imageRepository := repositories.NewImageRepository(pool)
	imageService := services.NewImageService(imageRepository)
	router := handlers.NewRouterWithProducts(supplierService, clientService, productService, imageService)

	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           router,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		ReadTimeout:       cfg.HTTPReadTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("HTTP server started", "address", cfg.HTTPAddress)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("serve HTTP: %w", err)
	case <-applicationContext.Done():
		slog.Info("shutting down HTTP server")
	}

	shutdownContext, cancelShutdown := context.WithTimeout(
		context.Background(),
		cfg.HTTPShutdownTimeout,
	)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownContext); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	if err := <-serverErrors; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP during shutdown: %w", err)
	}

	return nil
}
