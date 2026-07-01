package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"keeneye_practice/app/internal/db"
	"keeneye_practice/app/internal/handlers"
	"keeneye_practice/app/internal/middleware" // Новый импорт
	"keeneye_practice/app/internal/repository"
	"keeneye_practice/app/internal/service"
	"keeneye_practice/migrations"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		logger.Error("Failed to parse DB URL", "error", err)
		os.Exit(1)
	}

	dbPool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		logger.Error("Database connection failed", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	migrationDB, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Ошибка открытия БД для миграций: %v", err)
	}
	defer func() {
		if err := migrationDB.Close(); err != nil {
			logger.Error("Failed to close migration DB connection", "error", err)
		}
	}()

	goose.SetBaseFS(migrations.EmbedFS)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Ошибка установки диалекта goose: %v", err)
	}

	log.Println("Проверка и накат миграций...")
	if err := goose.Up(migrationDB, "."); err != nil {
		log.Fatalf("Ошибка применения миграций: %v", err)
	}
	log.Println("Миграции успешно применены!")

	queries := db.New(dbPool)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "super-secret-local-key"
	}

	authSvc := service.NewAuthService(queries, dbPool, jwtSecret)
	authHandler := handlers.NewAuthHandler(authSvc)
	studentRepo := repository.NewPostgresStudentRepository(queries)
	studentSvc := service.NewStudentService(studentRepo)
	studentHandler := handlers.NewStudentHandler(studentSvc)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Логирование HTTP запросов
	r.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		logger.Info("HTTP Request", "method", c.Request.Method, "path", path, "status", c.Writer.Status(), "latency", time.Since(start).String())
	})

	auth := r.Group("/api/auth")
	{
		auth.POST("/login", authHandler.Login)
	}

	api := r.Group("/api/base")
	api.Use(middleware.AuthMiddleware(jwtSecret))
	{
		api.GET("/students", middleware.RequireRoles("admin", "teacher", "student"), studentHandler.GetAll)
		api.GET("/students/:id", middleware.RequireRoles("admin", "teacher", "student"), studentHandler.GetByID)
		api.PUT("/students", middleware.RequireRoles("admin", "teacher", "student"), studentHandler.Update)

		api.DELETE("/students/:id", middleware.RequireRoles("admin"), studentHandler.Delete)
	}

	r.GET("/healthz", func(c *gin.Context) {
		if err := dbPool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "up"})
	})

	srv := &http.Server{Addr: ":8080", Handler: r}

	go func() {
		logger.Info("App server listening on port :8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Listen error", "error", err)
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Forced shutdown enforced", "error", err)
	}
	logger.Info("Application stopped cleanly")
}
