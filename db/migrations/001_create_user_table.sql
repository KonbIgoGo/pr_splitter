-- +goose Up

CREATE TABLE user_t
(
    id           TEXT PRIMARY KEY,
    user_name    TEXT NOT NULL,
    team_name    TEXT NOT NULL,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE
);

-- +goose Down
DROP TABLE user_t;