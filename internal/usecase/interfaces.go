package usecase

import (
	"context"

	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/KonbIgoGo/pr_splitter/internal/repository"
	"go.uber.org/zap"
)

type PRUseCase interface {
	PullRequestCreate(ctx context.Context, id string, name string, authorID string) (generated.PullRequest, error)
	PullRequestMerge(ctx context.Context, id string) (generated.PullRequest, error)
	PullRequestReassign(ctx context.Context, id string, oldUserID string) (generated.PullRequest, error)
}

type UserUseCase interface {
	UserSetIsActive(ctx context.Context, id string, isActive bool) (generated.User, error)
	UserGetReview(ctx context.Context, id string) ([]generated.PullRequest, error)
}

type TeamUseCase interface {
	TeamAdd(ctx context.Context, name string, memberIDs []string) (generated.Team, error)
	TeamGet(ctx context.Context, name string) (generated.Team, error)
}

var _ PRUseCase = (*useCaseImpl)(nil)
var _ UserUseCase = (*useCaseImpl)(nil)
var _ TeamUseCase = (*useCaseImpl)(nil)

type useCaseImpl struct {
	logger         *zap.Logger
	prRepository   repository.PRRepository
	userRepository repository.UserRepository
	teamRepository repository.TeamRepository
}

func New(
	logger *zap.Logger,
	prRepository repository.PRRepository,
	userRepository repository.UserRepository,
	teamRepository repository.TeamRepository,
) *useCaseImpl {
	return &useCaseImpl{
		logger:         logger,
		prRepository:   prRepository,
		userRepository: userRepository,
		teamRepository: teamRepository,
	}
}
