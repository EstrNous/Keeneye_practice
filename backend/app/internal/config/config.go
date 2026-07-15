package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type EmailRetryBackendType string

const (
	EmailRetryBackendWorker EmailRetryBackendType = "worker"
	EmailRetryBackendKafka  EmailRetryBackendType = "kafka"
)

type Config struct {
	DatabaseURL   string
	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration
	LogLevel      string
	SeedDevData   bool
	AppEnv        string

	RegistrationLinkTTL        time.Duration
	RegistrationBaseURL        string
	EmailRetryBackend          EmailRetryBackendType
	KafkaBrokers               []string
	KafkaEmailTopic            string
	KafkaConsumerGroup         string
	EmailRetryMaxAttempts      int32
	EmailRetrySimulateFailRate float64
	ExpirationWorkerInterval   time.Duration
	EmailRetryWorkerInterval   time.Duration

	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	MailerType   string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTAccessTTL:  15 * time.Minute,
		JWTRefreshTTL: 7 * 24 * time.Hour,
		LogLevel:      getenv("LOG_LEVEL", "info"),
		AppEnv:        getenv("APP_ENV", "development"),

		RegistrationLinkTTL:        24 * time.Hour,
		RegistrationBaseURL:        getenv("REGISTRATION_BASE_URL", "http://localhost"),
		EmailRetryBackend:          EmailRetryBackendWorker,
		KafkaEmailTopic:            getenv("KAFKA_EMAIL_TOPIC", "registration-email-retry"),
		KafkaConsumerGroup:         getenv("KAFKA_CONSUMER_GROUP", "registration-email"),
		EmailRetryMaxAttempts:      5,
		EmailRetrySimulateFailRate: 0.3,
		ExpirationWorkerInterval:   30 * time.Second,
		EmailRetryWorkerInterval:   15 * time.Second,
		MailerType:                 getenv("MAILER_TYPE", "log"),
		SMTPHost:                   os.Getenv("SMTP_HOST"),
		SMTPPort:                   getenv("SMTP_PORT", "587"),
		SMTPUser:                   os.Getenv("SMTP_USER"),
		SMTPPassword:               os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:                   getenv("SMTP_FROM", "noreply@localhost"),
	}

	if v := os.Getenv("JWT_ACCESS_TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
		}
		cfg.JWTAccessTTL = d
	}

	if v := os.Getenv("JWT_REFRESH_TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
		}
		cfg.JWTRefreshTTL = d
	}

	if v := os.Getenv("REGISTRATION_LINK_TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid REGISTRATION_LINK_TTL: %w", err)
		}
		cfg.RegistrationLinkTTL = d
	}

	if v := os.Getenv("EXPIRATION_WORKER_INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid EXPIRATION_WORKER_INTERVAL: %w", err)
		}
		cfg.ExpirationWorkerInterval = d
	}

	if v := os.Getenv("EMAIL_RETRY_WORKER_INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid EMAIL_RETRY_WORKER_INTERVAL: %w", err)
		}
		cfg.EmailRetryWorkerInterval = d
	}

	if v := os.Getenv("EMAIL_RETRY_MAX_ATTEMPTS"); v != "" {
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid EMAIL_RETRY_MAX_ATTEMPTS: %w", err)
		}
		cfg.EmailRetryMaxAttempts = int32(n)
	}

	if v := os.Getenv("EMAIL_RETRY_SIMULATE_FAILURE_RATE"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid EMAIL_RETRY_SIMULATE_FAILURE_RATE: %w", err)
		}
		cfg.EmailRetrySimulateFailRate = f
	}

	switch strings.ToLower(getenv("EMAIL_RETRY_BACKEND", "worker")) {
	case "kafka":
		cfg.EmailRetryBackend = EmailRetryBackendKafka
	default:
		cfg.EmailRetryBackend = EmailRetryBackendWorker
	}

	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		cfg.KafkaBrokers = strings.Split(brokers, ",")
	}

	cfg.SeedDevData = parseBool(os.Getenv("SEED_DEV_DATA"))

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.JWTSecret == "" {
		if cfg.AppEnv == "production" {
			return nil, fmt.Errorf("JWT_SECRET is required in production")
		}
		cfg.JWTSecret = "super-secret-local-key"
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseBool(v string) bool {
	b, _ := strconv.ParseBool(v)
	return b
}
