-- +goose Up

CREATE TABLE team
(
    team_name           TEXT PRIMARY KEY
);

-- +goose Down
DROP TABLE team;