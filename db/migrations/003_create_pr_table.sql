-- +goose Up

CREATE TABLE pr
(
    name                    TEXT NOT NULL,
    id                      TEXT PRIMARY KEY,
    author_id               TEXT NOT NULL,
    status                  TEXT NOT NULL
    CHECK (status IN ('OPEN', 'MERGED')),   
    created_at TIMESTAMP    DEFAULT now() NOT NULL,
    merged_at TIMESTAMP
);

-- +goose Down
DROP TABLE pr;