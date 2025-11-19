package repository

import (
	"context"

	"github.com/KonbIgoGo/pr_splitter/internal/entity"
)

type PRRepository interface {
	CreatePullRequest(ctx context.Context, id string, name string, authorID string) (entity.PR, error)
	MergePullRequest(ctx context.Context, id string) (entity.PR, error)
	ReassignPullRequest(ctx context.Context, id string, oldUserID string) (entity.PR, error)
}

type UserRepository interface {
	SetIsActiveUser(ctx context.Context, id string, isActive bool) (entity.User, error)
	GetReviewUser(ctx context.Context, id string) ([]entity.PR, error)
}

type TeamRepository interface {
	AddTeam(ctx context.Context, team entity.Team) (entity.Team, error)
	GetTeam(ctx context.Context, name string) (entity.Team, error)
}
