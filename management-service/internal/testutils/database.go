package testutils

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"

	"github.com/caraml-dev/xp/management-service/config"
	db "github.com/caraml-dev/xp/management-service/database"
)

const MigrationsPath = "file://../database/db-migrations"

var dbConfig config.DatabaseConfig = config.DatabaseConfig{
	Host:           GetEnvOrDefault("DATABASE_HOST", "localhost"),
	Port:           5432,
	User:           GetEnvOrDefault("DATABASE_USER", "xp"),
	Password:       GetEnvOrDefault("DATABASE_PASSWORD", "xp"),
	Database:       GetEnvOrDefault("DATABASE_NAME", "xp"),
	MigrationsPath: MigrationsPath,
}

// CreateTestDB connects to test postgreSQL instance (either local or the one
// at CI environment) and creates a new database with an up-to-date schema.
// It returns a reference to the DB and a clean up function if successful.
func CreateTestDB(migrationsPath string) (*gorm.DB, func(), error) {
	if migrationsPath != "" {
		dbConfig.MigrationsPath = migrationsPath
	}
	testDBCfg := dbConfig
	testDBCfg.Database = fmt.Sprintf("mlp_id_%d", time.Now().UnixNano())

	connStr := db.ConnectionString(&dbConfig)
	log.Printf("connecting to test db: %s", connStr)
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, nil, err
	}

	testDB, err := create(conn, &testDBCfg)
	if err != nil {
		return nil, nil, err
	}

	// Get underlying sql DB
	sqlDB, err := testDB.DB()
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		if err := sqlDB.Close(); err != nil {
			log.Fatalf("Failed to close connection to integration test database: \n%s", err)
		} else if _, err := conn.Exec("DROP DATABASE " + testDBCfg.Database); err != nil {
			log.Fatalf("Failed to cleanup integration test database: \n%s", err)
		} else if err = conn.Close(); err != nil {
			log.Fatalf("Failed to close database: \n%s", err)
		}
	}

	if err = db.Migrate(&testDBCfg); err != nil {
		cleanup()
		return nil, nil, err
	}

	return testDB, cleanup, nil
}

func create(conn *sql.DB, newDBCfg *config.DatabaseConfig) (*gorm.DB, error) {
	if _, err := conn.Exec("CREATE DATABASE " + newDBCfg.Database); err != nil {
		return nil, err
	} else if gormDB, err := db.Open(newDBCfg); err != nil {
		if _, err := conn.Exec("DROP DATABASE " + newDBCfg.Database); err != nil {
			log.Fatalf("Failed to cleanup integration test database: \n%s", err)
		}
		return nil, err
	} else {
		return gormDB, nil
	}
}
