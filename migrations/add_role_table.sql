CREATE TABLE IF NOT EXISTS role
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name       TEXT        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz DEFAULT now()
)