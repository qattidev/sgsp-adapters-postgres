-- +goose Up
CREATE TABLE IF NOT EXISTS sgsp_group_assignments (
    app_id text NOT NULL,
    group_key text NOT NULL,
    app_version text NOT NULL,
    owner_id text NOT NULL,
    incarnation bytea NOT NULL CHECK (octet_length(incarnation) = 16),
    endpoint_address text NOT NULL,
    endpoint_server_name text NOT NULL,
    closed boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    closed_at timestamptz,
    PRIMARY KEY (app_id, group_key),
    CHECK (closed = (closed_at IS NOT NULL))
);

CREATE TABLE IF NOT EXISTS sgsp_schema_migrations (
    version text PRIMARY KEY,
    checksum bytea NOT NULL,
    applied_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO sgsp_schema_migrations(version, checksum) VALUES ('001_group_assignments', decode('cc20847678bdc7fbbb336b4083e8baa59cfc036594b16abf462ce47495782c2f', 'hex')) ON CONFLICT (version) DO NOTHING;

-- +goose Down
DROP TABLE sgsp_schema_migrations;
DROP TABLE sgsp_group_assignments;
