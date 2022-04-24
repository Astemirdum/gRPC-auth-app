package server

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/Astemirdum/user-app/server/schema"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type ConfigDB struct {
	Host     string
	Port     int
	Username string
	Password string
	NameDB   string
}

func NewPostgresDB(cfg *ConfigDB) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", newDSN(cfg))
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	if err = MigrateSchema(cfg, schema.Label); err != nil {
		return nil, err
	}
	return db, nil
}

func newDSN(cfg *ConfigDB) string {
	return fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.Username, cfg.NameDB, cfg.Password)
}

func MigrateSchema(cfg *ConfigDB, label string) error {
	db, err := sql.Open("postgres", newDSN(cfg))
	if err != nil {
		return err
	}
	src, err := httpfs.New(http.FS(schema.MigrationFiles), ".")
	if err != nil {
		return err
	}
	targetInstance, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: label + "_migrations",
	})
	if err != nil {
		return fmt.Errorf("cannot create target db instance: %w", err)
	}

	m, err := migrate.NewWithInstance("<embed>", src, "postgres", targetInstance)
	if err != nil {
		return fmt.Errorf("cannot create migration instance: %w", err)
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrations failed: %w", err)
	}

	err = targetInstance.Close()
	if err != nil {
		return fmt.Errorf("failed to close target db instance: %w", err)
	}
	return nil
}
