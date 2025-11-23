package usecase

import (
	"context"

	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/KonbIgoGo/pr_splitter/internal/entity"
)

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
