package repository

import (
	"database/sql"
	"fmt"
	"github.com/Astemirdum/user-app/server/internal/config"
	"github.com/Astemirdum/user-app/server/migrations"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
)

func NewPostgresDB(cfg *config.DB) (*sqlx.DB, error) {
	if err := MigrateSchema(cfg); err != nil {
		return nil, err
	}
	db, err := sqlx.Open("pgx", newDSN(cfg))
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func newDSN(cfg *config.DB) string {
	return fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.Username, cfg.NameDB, cfg.Password)
}

func MigrateSchema(cfg *config.DB) error {
	dsn := newDSN(cfg)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		return fmt.Errorf("migrateSchema ping: %w", err)
	}

	goose.SetBaseFS(migrations.MigrationFiles)

	if err = goose.Up(db, "."); err != nil {
		return errors.Wrap(err, "goose run()")
	}
	return nil
}
