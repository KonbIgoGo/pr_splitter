-- +goose Up

CREATE TABLE pr_reviewer
(
    pr_id           TEXT NOT NULL REFERENCES pr (id) ON DELETE CASCADE,
    reviewer_id     TEXT NOT NULL REFERENCES user_t (id) ON DELETE CASCADE,
    PRIMARY KEY (pr_id, reviewer_id)
);

-- +goose Down
DROP TABLE pr_reviewer;