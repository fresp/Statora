package configs

import (
	"errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const envFilePath = ".env"

func SaveDatabaseConfig(engine, mongoURI, mongoDBName, sqlitePath string) error {
	engine = strings.ToLower(strings.TrimSpace(engine))
	if engine != "mongodb" && engine != "sqlite" {
		return errors.New("engine must be mongodb or sqlite")
	}

	envMap, err := godotenv.Read(envFilePath)
	if err != nil {
		envMap = map[string]string{}
	}

	envMap["DB_ENGINE"] = engine

	if strings.TrimSpace(sqlitePath) == "" {
		sqlitePath = "./data/statusforge.db"
	}
	envMap["SQLITE_PATH"] = sqlitePath

	if engine == "mongodb" {
		mongoURI = strings.TrimSpace(mongoURI)
		mongoDBName = strings.TrimSpace(mongoDBName)
		if mongoURI == "" || mongoDBName == "" {
			return errors.New("mongodb requires MONGO_URI and MONGO_DB_NAME")
		}
		envMap["MONGO_URI"] = mongoURI
		envMap["MONGO_DB_NAME"] = mongoDBName
	}

	if err := godotenv.Write(envMap, envFilePath); err != nil {
		return err
	}

	for key, value := range envMap {
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return nil
}
