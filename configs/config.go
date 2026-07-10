package configs

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv             string
	MongoURI           string
	MongoDBName        string
	DBDriver           string
	DBHost             string
	DBPort             string
	DBDatabase         string
	DBUsername         string
	DBPassword         string
	DBSSLMode          string
	DBPath             string
	RedisAddr          string
	JWTSecret          string
	EmailEncryptionKey string
	MFASecretKey       string
	Port               string
	AdminEmail         string
	AdminPass          string
	AdminUser          string
	EnableWorker       bool
	GracefulShutdown   bool
	GracefulTimeout    int
}

func Load() *Config {
	return &Config{
		AppEnv:             normalizeAppEnv(getEnv("APP_ENV", "development")),
		MongoURI:           getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDBName:        getEnv("MONGODB_DB", "statusplatform"),
		DBDriver:           getEnv("DB_DRIVER", "mongodb"),
		DBHost:             getEnv("DB_HOST", ""),
		DBPort:             getEnv("DB_PORT", ""),
		DBDatabase:         getEnv("DB_DATABASE", ""),
		DBUsername:         getEnv("DB_USERNAME", ""),
		DBPassword:         getEnv("DB_PASSWORD", ""),
		DBSSLMode:          getEnv("DB_SSLMODE", ""),
		DBPath:             getEnv("DB_PATH", ""),
		RedisAddr:          getEnv("REDIS_URI", "localhost:6379"),
		JWTSecret:          getEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production"),
		EmailEncryptionKey: getEnv("APP_ENCRYPTION_KEY", ""),
		MFASecretKey:       getEnv("MFA_SECRET_KEY", ""),
		Port:               getEnv("PORT", "8080"),
		AdminEmail:         getEnv("ADMIN_EMAIL", "admin@statusplatform.com"),
		AdminPass:          getEnv("ADMIN_PASSWORD", "admin123"),
		AdminUser:          getEnv("ADMIN_USERNAME", "admin"),
		EnableWorker:       getBoolEnv("ENABLE_WORKER", "true"),
		GracefulShutdown:   getBoolEnv("GRACEFUL_SHUTDOWN", "true"),
		GracefulTimeout:    getEnvInt("SHUTDOWN_TIMEOUT", 30),
	}
}

func (c *Config) Validate() error {
	if len(c.EmailEncryptionKey) != 32 {
		return fmt.Errorf("APP_ENCRYPTION_KEY must be exactly 32 bytes")
	}

	if c.isProductionLike() {
		placeholderFields := requiredPlaceholderFields(
			requiredField{name: "JWT_SECRET", value: c.JWTSecret},
			requiredField{name: "APP_ENCRYPTION_KEY", value: c.EmailEncryptionKey},
			requiredField{name: "ADMIN_PASSWORD", value: c.AdminPass},
		)

		switch strings.TrimSpace(c.DBDriver) {
		case "mongodb":
			placeholderFields = append(placeholderFields, requiredPlaceholderFields(requiredField{name: "MONGODB_URI", value: c.MongoURI})...)
		case "postgres", "mysql":
			placeholderFields = append(placeholderFields,
				requiredPlaceholderFields(
					requiredField{name: "DB_USERNAME", value: c.DBUsername},
					requiredField{name: "DB_PASSWORD", value: c.DBPassword},
				)...,
			)
		}

		if len(placeholderFields) > 0 {
			return fmt.Errorf("production configuration contains change-me placeholders: %s", strings.Join(placeholderFields, ", "))
		}
	}

	switch strings.TrimSpace(c.DBDriver) {
	case "mongodb":
		return nil
	case "postgres", "mysql":
		missing := requiredMissing(
			requiredField{name: "DB_HOST", value: c.DBHost},
			requiredField{name: "DB_PORT", value: c.DBPort},
			requiredField{name: "DB_DATABASE", value: c.DBDatabase},
			requiredField{name: "DB_USERNAME", value: c.DBUsername},
		)
		if strings.TrimSpace(c.DBPort) != "" && !validPort(c.DBPort) {
			missing = append(missing, "DB_PORT")
		}
		if len(missing) > 0 {
			return fmt.Errorf("missing required database configuration for %s: %s", strings.TrimSpace(c.DBDriver), strings.Join(missing, ", "))
		}
	case "sqlite":
		if strings.TrimSpace(c.DBPath) == "" {
			return fmt.Errorf("missing required database configuration for sqlite: DB_PATH")
		}
	default:
		return fmt.Errorf("unsupported DB_DRIVER; supported values: mongodb, postgres, mysql, sqlite")
	}

	return nil
}

type requiredField struct {
	name  string
	value string
}

func requiredMissing(fields ...requiredField) []string {
	missing := make([]string, 0, len(fields))
	for _, field := range fields {
		if strings.TrimSpace(field.value) == "" {
			missing = append(missing, field.name)
		}
	}
	return missing
}

func requiredPlaceholderFields(fields ...requiredField) []string {
	placeholders := make([]string, 0, len(fields))
	for _, field := range fields {
		if isPlaceholder(field.value) {
			placeholders = append(placeholders, field.name)
		}
	}
	return placeholders
}

func isPlaceholder(value string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	return trimmed != "" && strings.Contains(trimmed, "change-me")
}

func normalizeAppEnv(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return "development"
	}
	return trimmed
}

func (c *Config) isProductionLike() bool {
	return normalizeAppEnv(c.AppEnv) != "development"
}

func validPort(port string) bool {
	value, err := strconv.Atoi(strings.TrimSpace(port))
	return err == nil && value > 0 && value <= 65535
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return fallback
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		return fallback
	}

	return val
}

func getBoolEnv(key, fallback string) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback == "true"
	}
	return v == "true"
}
