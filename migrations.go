package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"errors"
	"io/fs"
	"qattidev/sgsp-postgres/internal/dbgen"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

//go:embed legacy/001_group_assignments.sql
var legacyMigration []byte

// ErrMigrationDrift preserves the original adapter's checksum guard when
// adopting an existing database into Goose's migration history.
var ErrMigrationDrift = errors.New("sgsp postgres: migration checksum mismatch")

var ErrInvalidDatabase = errors.New("sgsp postgres: nil database")

// ApplyMigrations explicitly applies the embedded Goose migrations. Run this
// once during deployment, before starting application instances. The caller
// owns the database and its connection configuration.
func ApplyMigrations(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrInvalidDatabase
	}
	q := dbgen.New(db)
	checksum := sha256.Sum256(legacyMigration)
	exists, err := q.HasLegacyHistory(ctx)
	if err != nil {
		return err
	}
	if exists {
		stored, err := q.LegacyChecksum(ctx)
		if err != nil {
			return err
		}
		if !bytes.Equal(stored, checksum[:]) {
			return ErrMigrationDrift
		}
	}
	source, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return err
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, source)
	if err != nil {
		return err
	}
	_, err = provider.Up(ctx)
	if err != nil {
		return err
	}
	return nil
}
