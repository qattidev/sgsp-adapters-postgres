# SGSP postgres assignment store

Independent Go module `qattidev/sgsp-postgres` implementing
`qattidev/sgsp/placement.AssignmentStore`. SGSP does not import this module.
The application owns the database pool, explicitly calls `ApplyMigrations`,
constructs `New(db)`, and injects the store into `placement.BootstrapConfig`.
`Store.Close` closes a group; the application calls `db.Close()` at shutdown.

## Development

This repository is checked out as the SGSP submodule `adapters/postgres`.
Its local `replace qattidev/sgsp => ../..` supports that unpublished layout.
For a standalone checkout, change that replacement to your SGSP checkout;
when publishing, replace the development dependency with a released SGSP version.
No database dependencies are added to the SGSP module.

```sh
go generate ./...  # pinned SQLC v1.30.0; generated code is committed
go test -race ./...
go vet ./...
make check-generated
```

Edit `queries/*.sql` and Goose migrations under `migrations/`, then regenerate.
SQLC is a development tool, not a runtime dependency. Runtime dependencies are
SGSP and Goose; the database driver below is only imported by tests and host
applications. Migrations are embedded and run through an instance-local Goose
provider (no global Goose configuration). Run `ApplyMigrations` once as a
serialized deployment step before starting application instances. `New` never
creates or migrates tables. Storage errors propagate without memory fallback.

Assignments are keyed by application ID and group key, with a version invariant.
Assignment never replaces a winner or reopens a closed group. Closure checks
both owner ID and incarnation in the transaction and commits before returning.
Tests cover independent pools and processes, competing owners, closure races, version and
owner rejection, cancellation, storage outages, reopening, and bootstrap owner
availability/restart behavior.

## PostgreSQL configuration

Use a `database/sql` driver such as `github.com/jackc/pgx/v5/stdlib`. Configure
`search_path` on **every** pooled connection (for pgx, use RuntimeParams on its
connection config). All tables and Goose history belong to that schema.
Assign uses INSERT ON CONFLICT followed by a separate SELECT at READ COMMITTED;
Close locks the record with SELECT FOR UPDATE before checking its owner.

```go
import (
    "context"
    "database/sql"
    _ "github.com/jackc/pgx/v5/stdlib"
    postgres "qattidev/sgsp-postgres"
)

// In application startup; handle each error before proceeding:
db, err := sql.Open("pgx", databaseURL)
err = postgres.ApplyMigrations(context.Background(), db)
store, err := postgres.New(db)
// Inject store into placement.BootstrapConfig.Store.
```

Set `SGSP_TEST_DATABASE_URL` to a disposable PostgreSQL database for tests.
Tests create/drop unique schemas; a missing URL fails instead of skipping.
`scripts/run-integration.sh` retains test evidence under `artifacts/`.

### Existing SGSP PostgreSQL databases

Change imports from `qattidev/sgsp/placement/postgres` to
`qattidev/sgsp-postgres`. ApplyMigrations verifies the original
`sgsp_schema_migrations` checksum before adopting the schema into Goose's
`goose_db_version`. Existing assignment data is preserved. The original SQL is
retained under `legacy/` for this checksum; do not edit it. The legacy drift
check remains; subsequent migration version tracking is owned by Goose.
Never run a Down migration on production assignment data: deleting closed
records permits forbidden group-key reuse. Schema deployments are serialized
by the application; runtime stores do not run migrations automatically.
