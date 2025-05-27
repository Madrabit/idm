-- +goose Up
CREATE TABLE IF NOT EXISTS employee
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name       TEXT        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz          DEFAULT now()
    );
CREATE TABLE IF NOT EXISTS role
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name       TEXT        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz DEFAULT now()
    );

-- +goose Down
DROP TABLE IF EXISTS employee;
DROP TABLE IF EXISTS role;
