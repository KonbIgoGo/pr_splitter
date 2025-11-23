package repository

import (
	"context"

	"github.com/KonbIgoGo/pr_splitter/internal/entity"
)

//go:generate go tool go.uber.org/mock/mockgen -source=./interfaces.go -destination=../../mocks/repositoryMock.go -package=mocks
type PRRepository interface {
	CreatePullRequest(ctx context.Context, id string, name string, authorID string) (entity.PR, error)
	MergePullRequest(ctx context.Context, id string) (entity.PR, error)
	ReassignPullRequest(ctx context.Context, id string, oldUserID string) (entity.PR, string, error)
}

type UserRepository interface {
	GetUsersReviewStatistic(ctx context.Context) ([]entity.Statistic, error)
	GetUsersPRAuthorityStatistic(ctx context.Context) ([]entity.Statistic, error)
	SetIsActiveUser(ctx context.Context, id string, isActive bool) (entity.User, error)
	GetReviewUser(ctx context.Context, id string) ([]entity.PR, error)
}

type TeamRepository interface {
	AddTeam(ctx context.Context, team entity.Team) (entity.Team, error)
	GetTeam(ctx context.Context, name string) (entity.Team, error)
}
