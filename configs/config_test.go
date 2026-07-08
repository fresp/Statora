package configs

import (
	"strings"
	"testing"
)

func TestLoadDatabaseDefaults(t *testing.T) {
	t.Setenv("DB_DRIVER", "")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_DATABASE", "")
	t.Setenv("DB_USERNAME", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_SSLMODE", "")
	t.Setenv("DB_PATH", "")
	t.Setenv("MONGODB_URI", "")
	t.Setenv("MONGODB_DB", "")

	cfg := Load()

	if cfg.DBDriver != "mongodb" {
		t.Fatalf("DBDriver = %q, want mongodb", cfg.DBDriver)
	}
	if cfg.MongoURI != "mongodb://localhost:27017" {
		t.Fatalf("MongoURI = %q, want default Mongo URI", cfg.MongoURI)
	}
	if cfg.MongoDBName != "statusplatform" {
		t.Fatalf("MongoDBName = %q, want statusplatform", cfg.MongoDBName)
	}
}

func TestLoadDatabaseEnvFields(t *testing.T) {
	t.Setenv("DB_DRIVER", "postgres")
	t.Setenv("DB_HOST", "db.example.test")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_DATABASE", "statora")
	t.Setenv("DB_USERNAME", "statora_user")
	t.Setenv("DB_PASSWORD", "secret-password")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("DB_PATH", "/tmp/statora.db")
	t.Setenv("MONGODB_URI", "mongodb://mongo.example.test:27017")
	t.Setenv("MONGODB_DB", "statora_mongo")

	cfg := Load()

	checks := map[string]string{
		"DBDriver":    cfg.DBDriver,
		"DBHost":      cfg.DBHost,
		"DBPort":      cfg.DBPort,
		"DBDatabase":  cfg.DBDatabase,
		"DBUsername":  cfg.DBUsername,
		"DBPassword":  cfg.DBPassword,
		"DBSSLMode":   cfg.DBSSLMode,
		"DBPath":      cfg.DBPath,
		"MongoURI":    cfg.MongoURI,
		"MongoDBName": cfg.MongoDBName,
	}
	wants := map[string]string{
		"DBDriver":    "postgres",
		"DBHost":      "db.example.test",
		"DBPort":      "5432",
		"DBDatabase":  "statora",
		"DBUsername":  "statora_user",
		"DBPassword":  "secret-password",
		"DBSSLMode":   "require",
		"DBPath":      "/tmp/statora.db",
		"MongoURI":    "mongodb://mongo.example.test:27017",
		"MongoDBName": "statora_mongo",
	}

	for field, got := range checks {
		if got != wants[field] {
			t.Fatalf("%s = %q, want %q", field, got, wants[field])
		}
	}
}

func TestLoadAllowedOriginsDefaultLocalDevelopment(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "")

	cfg := Load()

	wants := []string{
		"http://localhost:8080",
		"http://127.0.0.1:8080",
		"http://localhost:5173",
		"http://127.0.0.1:5173",
	}
	if strings.Join(cfg.AllowedOrigins, ",") != strings.Join(wants, ",") {
		t.Fatalf("AllowedOrigins = %v, want %v", cfg.AllowedOrigins, wants)
	}
}

func TestLoadAllowedOriginsFromEnv(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", " https://status.example.com, http://localhost:5173 ,,https://admin.example.com ")

	cfg := Load()

	wants := []string{"https://status.example.com", "http://localhost:5173", "https://admin.example.com"}
	if strings.Join(cfg.AllowedOrigins, ",") != strings.Join(wants, ",") {
		t.Fatalf("AllowedOrigins = %v, want %v", cfg.AllowedOrigins, wants)
	}
}

func TestConfigValidateDatabaseDriver(t *testing.T) {
	validKey := strings.Repeat("a", 32)
	tests := []struct {
		name       string
		cfg        Config
		wantErr    bool
		wantFields []string
	}{
		{
			name: "mongodb is valid by default",
			cfg: Config{
				EmailEncryptionKey: validKey,
				DBDriver:           "mongodb",
			},
		},
		{
			name: "unsupported driver fails without leaking input",
			cfg: Config{
				EmailEncryptionKey: validKey,
				DBDriver:           "secret-password",
			},
			wantErr:    true,
			wantFields: []string{"DB_DRIVER"},
		},
		{
			name: "sqlite requires path",
			cfg: Config{
				EmailEncryptionKey: validKey,
				DBDriver:           "sqlite",
			},
			wantErr:    true,
			wantFields: []string{"DB_PATH"},
		},
		{
			name: "sqlite accepts path",
			cfg: Config{
				EmailEncryptionKey: validKey,
				DBDriver:           "sqlite",
				DBPath:             "/tmp/statora.db",
			},
		},
		{
			name: "postgres requires connection fields",
			cfg: Config{
				EmailEncryptionKey: validKey,
				DBDriver:           "postgres",
				DBPassword:         "do-not-leak",
			},
			wantErr:    true,
			wantFields: []string{"DB_HOST", "DB_PORT", "DB_DATABASE", "DB_USERNAME"},
		},
		{
			name: "mysql requires connection fields",
			cfg: Config{
				EmailEncryptionKey: validKey,
				DBDriver:           "mysql",
				DBPassword:         "do-not-leak",
			},
			wantErr:    true,
			wantFields: []string{"DB_HOST", "DB_PORT", "DB_DATABASE", "DB_USERNAME"},
		},
		{
			name: "postgres accepts connection fields",
			cfg: Config{
				EmailEncryptionKey: validKey,
				DBDriver:           "postgres",
				DBHost:             "localhost",
				DBPort:             "5432",
				DBDatabase:         "statora",
				DBUsername:         "statora",
			},
		},
		{
			name: "postgres rejects blank host and non numeric port",
			cfg: Config{
				EmailEncryptionKey: validKey,
				DBDriver:           "postgres",
				DBHost:             " ",
				DBPort:             "not-a-port",
				DBDatabase:         "statora",
				DBUsername:         "statora",
			},
			wantErr:    true,
			wantFields: []string{"DB_HOST", "DB_PORT"},
		},
		{
			name: "mysql accepts connection fields",
			cfg: Config{
				EmailEncryptionKey: validKey,
				DBDriver:           "mysql",
				DBHost:             "localhost",
				DBPort:             "3306",
				DBDatabase:         "statora",
				DBUsername:         "statora",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
			if err == nil {
				return
			}

			msg := err.Error()
			for _, field := range tt.wantFields {
				if !strings.Contains(msg, field) {
					t.Fatalf("Validate() error = %q, want field %s", msg, field)
				}
			}
			if strings.Contains(msg, "do-not-leak") || strings.Contains(msg, "DB_PASSWORD") {
				t.Fatalf("Validate() error leaked password information: %q", msg)
			}
			if strings.Contains(msg, "secret-password") {
				t.Fatalf("Validate() error leaked unsupported driver value: %q", msg)
			}
		})
	}
}

func TestConfigValidatePreservesEncryptionKeyValidation(t *testing.T) {
	cfg := Config{EmailEncryptionKey: "short", DBDriver: "mongodb"}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want APP_ENCRYPTION_KEY error")
	}
	if !strings.Contains(err.Error(), "APP_ENCRYPTION_KEY must be exactly 32 bytes") {
		t.Fatalf("Validate() error = %q, want APP_ENCRYPTION_KEY length error", err.Error())
	}
}
