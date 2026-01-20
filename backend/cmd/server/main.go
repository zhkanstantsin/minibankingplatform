package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"minibankingplatform/internal/api"
	"minibankingplatform/internal/infrastructure"
	"minibankingplatform/internal/service"
	"minibankingplatform/pkg/jwt"
	"minibankingplatform/pkg/trm"
	"minibankingplatform/pkg/trm/pgxfactory"
)

// Config holds application configuration.
type Config struct {
	// Database
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string

	// Server
	ServerPort string

	// JWT
	JWTSecret   string
	JWTDuration time.Duration
}

func main() {
	// Load configuration from environment
	cfg := loadConfig()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to database
	pool, err := connectDB(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Create transaction factory
	txFactory, err := pgxfactory.New(ctx, pool)
	if err != nil {
		log.Fatalf("Failed to create transaction factory: %v", err)
	}

	// Create transaction manager
	txManager := trm.NewTransactionManager(txFactory)

	// Create injector for repositories
	injector := trm.NewInjector[infrastructure.DBTX](pool)

	// Create repositories
	usersRepo := infrastructure.NewUsersRepository(injector)
	accountsRepo := infrastructure.NewAccountsRepository(injector)
	transfersRepo := infrastructure.NewTransfersRepository(injector)
	exchangesRepo := infrastructure.NewExchangesRepository(injector)
	transactionsRepo := infrastructure.NewTransactionsRepository(injector)
	ledgerRepo := infrastructure.NewLedgerRepository(injector)

	// Create exchange rate provider (1 USD = 0.92 EUR)
	exchangeRateProvider := infrastructure.NewFixedExchangeRateProvider(decimal.NewFromFloat(0.92))

	// Create JWT token manager
	tokenManager := jwt.NewTokenManager(cfg.JWTSecret, cfg.JWTDuration)

	// Create application service
	svc := service.NewService(
		txManager,
		usersRepo,
		accountsRepo,
		transfersRepo,
		exchangesRepo,
		transactionsRepo,
		ledgerRepo,
		exchangeRateProvider,
		tokenManager,
	)

	// Create API handler
	handler := api.NewAPIHandler(svc)

	// Setup router
	router := chi.NewRouter()

	// Add standard middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Timeout(60 * time.Second))

	// Add CORS middleware for development
	router.Use(corsMiddleware)

	// Add JWT authentication middleware
	router.Use(api.AuthMiddleware(tokenManager))

	// Register OpenAPI handlers
	strictHandler := api.NewStrictHandler(handler, nil)
	api.HandlerFromMux(strictHandler, router)

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shutting down...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}

func loadConfig() Config {
	return Config{
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "bankuser"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "bankpass123"),
		PostgresDB:       getEnv("POSTGRES_DB", "minibankingdb"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		JWTSecret:        getEnv("JWT_SECRET", "your-super-secret-key-change-in-production"),
		JWTDuration:      24 * time.Hour,
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func connectDB(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDB,
	)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("parsing connection string: %w", err)
	}

	// Configure pool
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	log.Println("Connected to database")
	return pool, nil
}

// corsMiddleware adds CORS headers for development.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
