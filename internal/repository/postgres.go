package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/KonbIgoGo/pr_splitter/internal/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

var _ PRRepository = (*postgresRepository)(nil)
var _ UserRepository = (*postgresRepository)(nil)
var _ TeamRepository = (*postgresRepository)(nil)

type postgresRepository struct {
	db *pgxpool.Pool
}

func (p *postgresRepository) CreatePullRequest(ctx context.Context, id string, name string, authorID string) (entity.PR, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return entity.PR{}, err
	}
	defer func() {
		err = tx.Rollback(ctx)
		if err != nil {
			logrus.Error(err)
		}
	}()

	const getReviewersAndCheckAuthor = /* sql */ `
		WITH author AS (
			SELECT id
			FROM user_t
			WHERE id = $1
		),
		reviewers AS (
			SELECT id
			FROM user_t
			WHERE id <> $1
				AND is_active = TRUE
				AND team_name = (SELECT team_name FROM author)
			ORDER BY random()
			LIMIT 2
		)
        SELECT *
        FROM author
        UNION ALL
		SELECT *
		FROM reviewers
	`

	rows, err := tx.Query(ctx, getReviewersAndCheckAuthor, authorID)
	if err != nil {
		return entity.PR{}, err
	}
	fmt.Println("collected")

	ids, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (string, error) {
		if err = row.Scan(&id); err != nil {
			return "", err
		}
		return id, nil
	})
	if err != nil {
		return entity.PR{}, err
	}

	if len(ids) == 0 {
		return entity.PR{}, entity.ErrUserNotFound
	}

	const insertPrData = /* sql */ `
		WITH new_pr AS (
			INSERT INTO pr (id, name, author_id, status, created_at)
			VALUES ($1, $2, $3, 'OPEN', now())
			RETURNING id, created_at, merged_at
		),
		ins_reviewers AS (
			INSERT INTO pr_reviewer (pr_id, reviewer_id)
			SELECT
				new_pr.id,
				r.reviewer_id
			FROM new_pr
			CROSS JOIN UNNEST($4::text[]) AS r(reviewer_id)
		)
		SELECT created_at
		FROM new_pr;
	`

	res := entity.PR{
		ID:                  id,
		Name:                name,
		AuthorID:            authorID,
		AssignedReviewersID: ids[1:],
		Status:              entity.OPEN,
	}

	err = tx.QueryRow(ctx, insertPrData, id, name, authorID, ids[1:]).Scan(&res.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return entity.PR{}, entity.ErrPRAlreadyExists
			}
		}

		return entity.PR{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return entity.PR{}, err
	}
	return res, nil
}
func (p *postgresRepository) MergePullRequest(ctx context.Context, id string) (entity.PR, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return entity.PR{}, err
	}
	defer func() {
		err = tx.Rollback(ctx)
		if err != nil {
			logrus.Error(err)
		}
	}()

	const mergePRQuery = /* sql */ `
		WITH update_pr AS
		(
			UPDATE pr
			SET
				status = 'MERGED',
				merged_at = COALESCE(merged_at, now())
			WHERE id = $1
			RETURNING name, author_id, created_at, merged_at
		),
		get_reviewers AS
		(
			SELECT ARRAY_AGG(reviewer_id) AS reviewer_ids FROM pr_reviewer
			WHERE pr_id = $1
		)
		SELECT
			u.name,
			u.author_id,
			u.created_at,
			u.merged_at,
			COALESCE(g.reviewer_ids, ARRAY[]::text[]) AS reviewer_ids
		FROM update_pr u
		LEFT JOIN get_reviewers g ON true;
	`

	var res entity.PR = entity.PR{
		Status: entity.MERGED,
		ID:     id,
	}

	err = tx.QueryRow(ctx, mergePRQuery, id).Scan(
		&res.Name,
		&res.AuthorID,
		&res.CreatedAt,
		&res.MergedAt,
		&res.AssignedReviewersID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.PR{}, entity.ErrPRNotFound
	}
	if err != nil {
		return entity.PR{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return entity.PR{}, err
	}
	return res, nil
}
func (p *postgresRepository) ReassignPullRequest(ctx context.Context, id string, oldUserID string) (entity.PR, string, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return entity.PR{}, "", err
	}
	defer func() {
		err = tx.Rollback(ctx)
		if err != nil {
			logrus.Error(err)
		}
	}()

	const checkQuery = /* sql */ `
		WITH pr_row AS (
			SELECT id, name, author_id, status, created_at, merged_at
			FROM pr
			WHERE id = $1
			FOR UPDATE
		),
		user_row AS (
			SELECT id
			FROM user_t
			WHERE id = $2
		),
		assigned AS (
			SELECT 1
			FROM pr_reviewer
			WHERE pr_id = $1 AND reviewer_id = $2
		)
		SELECT
			p.id,
			p.name,
			p.author_id,
			p.status,
			p.created_at,
			p.merged_at,
			EXISTS (SELECT 1 FROM user_row)   AS user_exists,
			EXISTS (SELECT 1 FROM assigned)   AS is_assigned
		FROM pr_row p;
	`

	var (
		prID, name, authorID string
		status               string
		createdAt            time.Time
		mergedAt             *time.Time
		userExists           bool
		isAssigned           bool
	)

	err = tx.QueryRow(ctx, checkQuery, id, oldUserID).Scan(
		&prID, &name, &authorID, &status, &createdAt, &mergedAt,
		&userExists, &isAssigned,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return entity.PR{}, "", entity.ErrPRNotFound
	}
	if err != nil {
		return entity.PR{}, "", err
	}
	if !userExists {
		return entity.PR{}, "", entity.ErrUserNotFound
	}
	//nolint:goconst // used in map
	if status == "MERGED" {
		return entity.PR{}, "", entity.ErrPRMerged
	}
	if !isAssigned {
		return entity.PR{}, "", entity.ErrReviewerIsNotAssignedToPR
	}

	const reassignPrQuery = /* sql */ `
		WITH author AS (
			SELECT author_id
			FROM pr
			WHERE id = $1
		),
		author_team AS (
			SELECT team_name
			FROM user_t
			WHERE id = (SELECT author_id FROM author)
		),
		candidate AS (
			SELECT u.id
			FROM user_t u
			WHERE u.is_active = TRUE
			AND u.team_name = (SELECT team_name FROM author_team)
			AND u.id <> $2
			AND u.id <> (SELECT author_id FROM author)
			AND NOT EXISTS (
				SELECT 1
				FROM pr_reviewer r
				WHERE r.pr_id = $1
					AND r.reviewer_id = u.id
			)
			ORDER BY random()
			LIMIT 1
		),
		update_reviewer AS (
			UPDATE pr_reviewer pr
			SET reviewer_id = c.id
			FROM candidate c
			WHERE pr.pr_id = $1
			AND pr.reviewer_id = $2
		),
		pr_row AS (
			SELECT id, name, author_id, status, created_at, merged_at
			FROM pr
			WHERE id = $1
		),
		reviewers AS (
			SELECT ARRAY_AGG(reviewer_id) AS reviewer_ids
			FROM pr_reviewer
			WHERE pr_id = $1
		)
		SELECT
			p.id,
			p.name,
			p.author_id,
			p.status,
			p.created_at,
			p.merged_at,
			r.reviewer_ids,
			c.id AS replaced_by
		FROM pr_row p
		JOIN reviewers r ON TRUE
		JOIN candidate c ON TRUE;
	`

	var pr entity.PR
	var replacedBy string

	err = tx.QueryRow(ctx, reassignPrQuery, id, oldUserID).Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&status,
		&pr.CreatedAt,
		&mergedAt,
		&pr.AssignedReviewersID,
		&replacedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.PR{}, "", entity.ErrPRNoCandidates
	}

	if err != nil {
		return entity.PR{}, "", err
	}

	pr.Status = mapStatus(status)

	if mergedAt != nil {
		pr.MergedAt = *mergedAt
	}

	err = tx.Commit(ctx)
	if err != nil {
		return entity.PR{}, "", err
	}
	return pr, replacedBy, nil
}

func (p *postgresRepository) SetIsActiveUser(ctx context.Context, id string, isActive bool) (entity.User, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return entity.User{}, err
	}
	defer func() {
		err = tx.Rollback(ctx)
		if err != nil {
			logrus.Error(err)
		}
	}()

	const updateActiveQuery = /* sql */ `
		UPDATE user_t
		SET
			is_active = $1
		WHERE id = $2
		RETURNING user_name, team_name;
	`

	res := entity.User{
		ID:       id,
		IsActive: isActive,
	}

	err = tx.QueryRow(ctx, updateActiveQuery, isActive, id).Scan(&res.Name, &res.TeamName)
	if err != nil {
		return entity.User{}, err
	}

	switch {
	case err == pgx.ErrNoRows:
		return entity.User{}, entity.ErrUserNotFound
	case err != nil:
		return entity.User{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return entity.User{}, err
	}
	return res, nil
}
func (p *postgresRepository) GetReviewUser(ctx context.Context, id string) ([]entity.PR, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = tx.Rollback(ctx)
		if err != nil {
			logrus.Error(err)
		}
	}()

	const getUserReviewsQuery = /* sql */ `
		WITH get_pr_ids AS (
			SELECT DISTINCT pr_id
			FROM pr_reviewer
			WHERE reviewer_id = $1
		),
		get_reviewers AS (
			SELECT
				pr_id,
				ARRAY_AGG(reviewer_id) AS reviewer_ids
			FROM pr_reviewer
			WHERE pr_id IN (SELECT pr_id FROM get_pr_ids)
			GROUP BY pr_id
		)
		SELECT
			p.id,
			p.name,
			p.author_id,
			p.status,
			p.created_at,
			p.merged_at,
			r.reviewer_ids
		FROM pr p
		JOIN get_pr_ids g ON g.pr_id = p.id
		LEFT JOIN get_reviewers r ON r.pr_id = p.id;
	`

	rows, err := tx.Query(ctx, getUserReviewsQuery, id)

	prs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (entity.PR, error) {
		var p entity.PR
		var status string
		var mergedAt sql.NullTime

		if err = row.Scan(
			&p.ID,
			&p.Name,
			&p.AuthorID,
			&status,
			&p.CreatedAt,
			&mergedAt,
			&p.AssignedReviewersID,
		); err != nil {
			return entity.PR{}, err
		}

		if mergedAt.Valid {
			t := mergedAt.Time
			p.MergedAt = t
		}
		p.Status = mapStatus(status)

		return p, nil
	})

	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}
	return prs, nil
}

func (p *postgresRepository) AddTeam(ctx context.Context, team entity.Team) (entity.Team, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return entity.Team{}, err
	}
	defer func() {
		err = tx.Rollback(ctx)
		if err != nil {
			logrus.Error(err)
		}
	}()

	const checkIfTeamExists = /* sql */ `
		SELECT 1
		FROM team
		WHERE team_name = $1
		LIMIT 1;
	`

	var dummy int
	err = tx.QueryRow(ctx, checkIfTeamExists, team.TeamName).Scan(&dummy)
	if err == nil {
		return entity.Team{}, entity.ErrTeamAlreadyExists
	}
	if err != pgx.ErrNoRows {
		return entity.Team{}, err
	}

	const insertTeamAndUsersQuery = /* sql */ `
		WITH new_team AS (
			INSERT INTO team (team_name)
			VALUES ($1)
			RETURNING team_name
		),
		upsert_users AS (
			INSERT INTO user_t (id, user_name, team_name, is_active)
			SELECT
				v.id,
				v.user_name,
				nt.team_name,
				v.is_active
			FROM (VALUES %s) AS v(id, user_name, is_active)
			CROSS JOIN new_team nt
			ON CONFLICT (id) DO UPDATE
			SET
				user_name = EXCLUDED.user_name,
				team_name = EXCLUDED.team_name,
				is_active = EXCLUDED.is_active
		)
		SELECT 1;
		`

	if len(team.Members) > 0 {
		valueStrings := make([]string, len(team.Members))
		args := make([]any, 0, 1+len(team.Members)*3)

		args = append(args, team.TeamName)

		for i, member := range team.Members {
			//nolint:mnd // defined indexes
			base := 2 + i*3
			//nolint:mnd // defined indexes
			valueStrings[i] = fmt.Sprintf("($%d,$%d,$%d::boolean)", base, base+1, base+2)
			args = append(args,
				member.UserID,
				member.Username,
				member.IsActive,
			)
		}

		query := fmt.Sprintf(insertTeamAndUsersQuery, strings.Join(valueStrings, ","))
		if _, err = tx.Exec(ctx, query, args...); err != nil {
			return entity.Team{}, err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return entity.Team{}, err
	}

	res := entity.Team{
		TeamName: team.TeamName,
		Members:  team.Members,
	}

	return res, nil
}

func (p *postgresRepository) GetTeam(ctx context.Context, name string) (team entity.Team, err error) {
	const checkTeamQuery = /* sql */ `
        SELECT 1
        FROM team
        WHERE team_name = $1
        FOR UPDATE
    `
	var dummy int
	if err = p.db.QueryRow(ctx, checkTeamQuery, name).Scan(&dummy); err != nil {
		if err == pgx.ErrNoRows {
			return entity.Team{}, entity.ErrTeamNotFound
		}
		return entity.Team{}, err
	}

	const getTeamQuery = /* sql */ `
        SELECT id, user_name, is_active
        FROM user_t
        WHERE team_name = $1
        FOR UPDATE
    `
	rows, err := p.db.Query(ctx, getTeamQuery, name)
	if err != nil {
		return entity.Team{}, err
	}

	members, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (entity.TeamMember, error) {
		var m entity.TeamMember
		if err = row.Scan(&m.UserID, &m.Username, &m.IsActive); err != nil {
			return entity.TeamMember{}, err
		}
		return m, nil
	})

	if err != nil {
		return entity.Team{}, err
	}

	team = entity.Team{
		TeamName: name,
		Members:  members,
	}

	return team, nil
}

func mapStatus(status string) entity.Status {
	if status == "MERGED" {
		return entity.MERGED
	}
	return entity.OPEN
}

func NewPostgresRepository(db *pgxpool.Pool) *postgresRepository {
	return &postgresRepository{db: db}
}
