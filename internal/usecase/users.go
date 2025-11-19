package usecase

import (
	"context"

	"github.com/KonbIgoGo/pr_splitter/generated"
)

func (i *useCaseImpl) UserSetIsActive(ctx context.Context, id string, isActive bool) (generated.User, error) {

}
func (i *useCaseImpl) UserGetReview(ctx context.Context, id string) ([]generated.PullRequest, error) {

}
