package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"keeneye_practice/app/internal/config"
	"keeneye_practice/app/internal/db"
	"keeneye_practice/app/internal/handlers"
	"keeneye_practice/app/internal/middleware"
	"keeneye_practice/app/internal/repository"
	"keeneye_practice/app/internal/seed"
	"keeneye_practice/app/internal/service"
	"keeneye_practice/app/internal/validators"
	"keeneye_practice/migrations"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parseLogLevel(cfg.LogLevel)}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to parse database url", "error", err)
		os.Exit(1)
	}

	dbPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	if err := runMigrations(cfg.DatabaseURL); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	queries := db.New(dbPool)

	authSvc := service.NewAuthService(queries, dbPool, cfg)
	groupRepo := repository.NewPostgresGroupRepository(queries)
	studentRepo := repository.NewPostgresStudentRepository(queries, dbPool)
	teacherRepo := repository.NewPostgresTeacherRepository(queries, dbPool)

	studentSvc := service.NewStudentService(studentRepo, authSvc, groupRepo)
	teacherSvc := service.NewTeacherService(teacherRepo)
	groupSvc := service.NewGroupService(groupRepo)

	if cfg.SeedDevData {
		if err := seed.DevData(ctx, queries, authSvc); err != nil {
			logger.Error("dev seed failed", "error", err)
			os.Exit(1)
		}
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := validators.RegisterCustomValidations(v); err != nil {
			logger.Error("failed to register custom validations", "error", err)
			os.Exit(1)
		}
	}

	authHandler := handlers.NewAuthHandler(authSvc)
	studentHandler := handlers.NewStudentHandler(studentSvc)
	teacherHandler := handlers.NewTeacherHandler(teacherSvc)
	groupHandler := handlers.NewGroupHandler(groupSvc)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	_ = r.SetTrustedProxies([]string{"127.0.0.1", "10.0.0.0/8", "172.16.0.0/12"})
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.InstanceID())
	r.Use(middleware.PrometheusMetrics())
	r.Use(httpLogMiddleware(logger))
	r.Use(middleware.ErrorHandler())

	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
		auth.POST("/refresh", authHandler.Refresh)
	}

	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		students := api.Group("/students")
		{
			students.GET("", middleware.RequireRoles("admin", "teacher", "student"), studentHandler.GetAll)
			students.GET("/:id", middleware.RequireRoles("admin", "teacher", "student"), studentHandler.GetByID)
			students.POST("", middleware.RequireRoles("admin"), studentHandler.Create)
			students.PUT("/:id", middleware.RequireRoles("admin", "teacher", "student"), studentHandler.Update)
			students.DELETE("/:id", middleware.RequireRoles("admin"), studentHandler.Delete)
		}

		teachers := api.Group("/teachers")
		{
			teachers.GET("", middleware.RequireRoles("admin"), teacherHandler.List)
			teachers.GET("/:id", middleware.RequireRoles("admin", "teacher"), teacherHandler.GetByID)
			teachers.PUT("/:id", middleware.RequireRoles("admin", "teacher"), teacherHandler.Update)
			teachers.DELETE("/:id", middleware.RequireRoles("admin"), teacherHandler.Delete)
			teachers.GET("/:id/groups", middleware.RequireRoles("admin", "teacher"), teacherHandler.ListGroups)
			teachers.POST("/:id/groups/:groupId", middleware.RequireRoles("admin"), teacherHandler.AssignGroup)
			teachers.DELETE("/:id/groups/:groupId", middleware.RequireRoles("admin"), teacherHandler.RemoveGroup)
		}

		groups := api.Group("/groups")
		{
			groups.GET("", middleware.RequireRoles("admin", "teacher"), groupHandler.List)
			groups.POST("", middleware.RequireRoles("admin"), groupHandler.Create)
		}
	}

	r.GET("/healthz", func(c *gin.Context) {
		if err := dbPool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "up"})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{Addr: ":8080", Handler: r}

	go func() {
		logger.Info("server listening", "addr", ":8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen error", "error", err)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("forced shutdown", "error", err)
	}
}

func runMigrations(dbURL string) error {
	migrationDB, err := sql.Open("pgx", dbURL)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := migrationDB.Close(); cerr != nil {
			slog.Error("migration database close failed", "error", cerr)
		}
	}()

	goose.SetBaseFS(migrations.EmbedFS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(migrationDB, ".")
}

func httpLogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		logger.Info("http request",
			"instance", c.GetString("instanceID"),
			"request_id", c.GetString("requestID"),
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency", time.Since(start).String(),
		)
	}
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
