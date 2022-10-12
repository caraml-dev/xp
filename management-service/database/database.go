package database

import (
	"errors"
	"fmt"
	"time"

	gomigrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/caraml-dev/xp/management-service/config"
)

var UtcLoc, _ = time.LoadLocation("UTC")

func ConnectionString(cfg *config.DatabaseConfig) string {
	return fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable TimeZone=UTC",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Database,
		cfg.Password)
}

func Open(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	return gorm.Open(pg.Open(ConnectionString(cfg)),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
}

func Migrate(cfg *config.DatabaseConfig) error {
	db, err := Open(cfg)
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
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
