-- +goose Up
CREATE TABLE repositories (
 owner text NOT NULL,
 name text NOT NULL,
 description text NOT NULL,
 stars bigint NOT NULL,
 forks bigint NOT NULL,
 created_at timestamptz NOT NULL,
 fetched_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE UNIQUE INDEX repositories_owner_name ON repositories (lower(owner), lower(name));

-- +goose Down
DROP TABLE repositories;
