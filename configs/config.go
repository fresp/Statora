package configs

import (
	"os"
	"strings"
)

type Config struct {
	MongoURI     string
	MongoDBName  string
	DBEngine     string
	SQLitePath   string
	SetupDone    bool
	RedisAddr    string
	JWTSecret    string
	MFASecretKey string
	Port         string
	AdminEmail   string
	AdminPass    string
	AdminUser    string
	EnableWorker bool
}

func Load() *Config {
	mongoURI, hasMongoURI := lookupNonEmptyEnv("MONGO_URI")
	mongoDBName, hasMongoDBName := lookupNonEmptyEnv("MONGO_DB_NAME")
	dbEngine, hasDBEngine := lookupNonEmptyEnv("DB_ENGINE")
	sqlitePath, hasSQLitePath := lookupNonEmptyEnv("SQLITE_PATH")

	if !hasDBEngine {
		switch {
		case hasMongoURI && hasMongoDBName:
			dbEngine = "mongodb"
		case hasSQLitePath:
			dbEngine = "sqlite"
		default:
			dbEngine = ""
		}
	}

	if !hasSQLitePath {
		sqlitePath = "./data/statusforge.db"
	}

	return &Config{
		MongoURI:     mongoURI,
		MongoDBName:  mongoDBName,
		DBEngine:     strings.ToLower(dbEngine),
		SQLitePath:   sqlitePath,
		SetupDone:    dbEngine != "",
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:    getEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production"),
		MFASecretKey: getEnv("MFA_SECRET_KEY", ""),
		Port:         getEnv("PORT", "8080"),
		AdminEmail:   getEnv("ADMIN_EMAIL", "admin@statusplatform.com"),
		AdminPass:    getEnv("ADMIN_PASSWORD", "admin123"),
		AdminUser:    getEnv("ADMIN_USERNAME", "admin"),
		EnableWorker: getBoolEnv("ENABLE_WORKER", "true"),
	}
}

func lookupNonEmptyEnv(key string) (string, bool) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return "", false
	}
	v = strings.TrimSpace(v)
	if v == "" {
		return "", false
	}
	return v, true
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getBoolEnv(key, fallback string) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback == "true"
	}
	return v == "true"
}
