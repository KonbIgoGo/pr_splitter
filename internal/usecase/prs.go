package usecase

import (
	"context"

	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/KonbIgoGo/pr_splitter/internal/entity"
)

func (i *useCaseImpl) PullRequestCreate(ctx context.Context, id string, name string, authorID string) (generated.PullRequest, error) {
	pr, err := i.prRepository.CreatePullRequest(ctx, id, name, authorID)
	if err != nil {
		return generated.PullRequest{}, err
	}

	status := generated.PullRequestShortStatusOPEN
	if pr.Status == entity.MERGED {
		status = generated.PullRequestShortStatusMERGED
	}

	return generated.PullRequest{
		AssignedReviewers: pr.AssignedReviewersID,
		AuthorId:          pr.AuthorID,
		CreatedAt:         &pr.CreatedAt,
		MergedAt:          &pr.MergedAt,
		PullRequestId:     pr.ID,
		PullRequestName:   pr.Name,
		Status:            generated.PullRequestStatus(status),
	}, nil
}
func (i *useCaseImpl) PullRequestMerge(ctx context.Context, id string) (generated.PullRequest, error) {
	pr, err := i.prRepository.MergePullRequest(ctx, id)
	if err != nil {
		return generated.PullRequest{}, err
	}

	status := generated.PullRequestShortStatusOPEN
	if pr.Status == entity.MERGED {
		status = generated.PullRequestShortStatusMERGED
	}

	return generated.PullRequest{
		AssignedReviewers: pr.AssignedReviewersID,
		AuthorId:          pr.AuthorID,
		CreatedAt:         &pr.CreatedAt,
		MergedAt:          &pr.MergedAt,
		PullRequestId:     pr.ID,
		PullRequestName:   pr.Name,
		Status:            generated.PullRequestStatus(status),
	}, nil
}
func (i *useCaseImpl) PullRequestReassign(ctx context.Context, id string, oldUserID string) (generated.PullRequest, error) {
	pr, err := i.prRepository.ReassignPullRequest(ctx, id, oldUserID)
	if err != nil {
		return generated.PullRequest{}, err
	}

	status := generated.PullRequestShortStatusOPEN
	if pr.Status == entity.MERGED {
		status = generated.PullRequestShortStatusMERGED
	}

	return generated.PullRequest{
		AssignedReviewers: pr.AssignedReviewersID,
		AuthorId:          pr.AuthorID,
		CreatedAt:         &pr.CreatedAt,
		MergedAt:          &pr.MergedAt,
		PullRequestId:     pr.ID,
		PullRequestName:   pr.Name,
		Status:            generated.PullRequestStatus(status),
	}, nil
}
