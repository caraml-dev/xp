package database

import (
	"errors"
	"fmt"

	gomigrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jinzhu/gorm"

	// Gorm requires this for interfacing with the postgres DB
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/caraml-dev/xp/management-service/config"
)

func ConnectionString(cfg *config.DatabaseConfig) string {
	return fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable timezone=UTC",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Database,
		cfg.Password)
}

func Open(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	return gorm.Open("postgres", ConnectionString(cfg))
}

func Migrate(cfg *config.DatabaseConfig) error {
	db, err := Open(cfg)
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db.DB(), &postgres.Config{})
	if err != nil {
		return err
	}
	defer driver.Close()

	if migrations, err := gomigrate.NewWithDatabaseInstance(cfg.MigrationsPath, cfg.Database, driver); err != nil {
		return err
	} else if err = migrations.Up(); err != nil && !errors.Is(err, gomigrate.ErrNoChange) {
		return err
	}
	return nil
}
