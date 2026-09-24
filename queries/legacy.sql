-- name: HasLegacyHistory :one
SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=current_schema() AND table_name='sgsp_schema_migrations');

-- name: LegacyChecksum :one
SELECT checksum FROM sgsp_schema_migrations WHERE version='001_group_assignments';

