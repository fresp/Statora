package configs

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
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
	AllowedOrigins     []string
}

var defaultAllowedOrigins = []string{
	"http://localhost:8080",
	"http://127.0.0.1:8080",
	"http://localhost:5173",
	"http://127.0.0.1:5173",
}

func Load() *Config {
	return &Config{
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
		AllowedOrigins:     getCSVEnv("ALLOWED_ORIGINS", defaultAllowedOrigins),
	}
}

func (c *Config) Validate() error {
	if len(c.EmailEncryptionKey) != 32 {
		return fmt.Errorf("APP_ENCRYPTION_KEY must be exactly 32 bytes")
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

func (c *Config) CORSAllowedOrigins() []string {
	if len(c.AllowedOrigins) == 0 {
		return append([]string(nil), defaultAllowedOrigins...)
	}
	return append([]string(nil), c.AllowedOrigins...)
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

func getCSVEnv(key string, fallback []string) []string {
	raw := os.Getenv(key)
	if strings.TrimSpace(raw) == "" {
		return append([]string(nil), fallback...)
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}
