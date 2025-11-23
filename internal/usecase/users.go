package usecase

import (
	"context"

	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/KonbIgoGo/pr_splitter/internal/entity"
)

func (i *useCaseImpl) UsersGetReviewStatistic(ctx context.Context) ([]generated.UserPRAuthorityStatistic, error) {
	stats, err := i.userRepository.GetUsersReviewStatistic(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]generated.UserPRAuthorityStatistic, 0, len(stats))
	for _, s := range stats {
		res = append(res, generated.UserPRAuthorityStatistic{PrsCount: s.Data, UserId: s.UserID})
	}

	return res, nil
}
func (i *useCaseImpl) UsersGetPRAuthorityStatistic(ctx context.Context) ([]generated.UserPRAuthorityStatistic, error) {
	stats, err := i.userRepository.GetUsersPRAuthorityStatistic(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]generated.UserPRAuthorityStatistic, 0, len(stats))
	for _, s := range stats {
		res = append(res, generated.UserPRAuthorityStatistic{PrsCount: s.Data, UserId: s.UserID})
	}

	return res, nil
}

func (i *useCaseImpl) UserSetIsActive(ctx context.Context, id string, isActive bool) (generated.User, error) {
	user, err := i.userRepository.SetIsActiveUser(ctx, id, isActive)
	if err != nil {
		return generated.User{}, err
	}

	return generated.User{
		IsActive: user.IsActive,
		TeamName: user.TeamName,
		UserId:   user.ID,
		Username: user.Name,
	}, nil
}
func (i *useCaseImpl) UserGetReview(ctx context.Context, id string) ([]generated.PullRequestShort, error) {
	reviews, err := i.userRepository.GetReviewUser(ctx, id)
	if err != nil {
		return nil, err
	}

	res := make([]generated.PullRequestShort, 0, len(reviews))
	for _, p := range reviews {
		status := generated.PullRequestShortStatusOPEN
		if p.Status == entity.MERGED {
			status = generated.PullRequestShortStatusMERGED
		}

		res = append(res, generated.PullRequestShort{
			AuthorId:        p.AuthorID,
			PullRequestId:   p.ID,
			PullRequestName: p.Name,
			Status:          status,
		})
	}

	return res, nil
}
