package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/KonbIgoGo/pr_splitter/internal/entity"
	"github.com/KonbIgoGo/pr_splitter/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPRUseCases(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	ctrl := gomock.NewController(t)
	prRepo := mocks.NewMockPRRepository(ctrl)
	usecase := New(nil, prRepo, nil, nil)

	type testCase struct {
		id          string
		oldUserID   string
		status      entity.Status
		name        string
		authorID    string
		replaced    string
		errExpected bool
	}

	t.Run("pr create", func(t *testing.T) {
		t.Parallel()
		tcs := []testCase{
			{
				id:          "valid",
				name:        "valid",
				authorID:    "authorValid",
				errExpected: false,
			},
			{
				id:          "invalid",
				name:        "invalid",
				authorID:    "authorInvalid",
				errExpected: true,
			},
		}

		for _, tc := range tcs {
			if tc.errExpected {
				prRepo.EXPECT().CreatePullRequest(ctx, tc.id, tc.name, tc.authorID).Return(entity.PR{}, errors.New("some error"))
			} else {
				prRepo.EXPECT().CreatePullRequest(ctx, tc.id, tc.name, tc.authorID).Return(entity.PR{
					ID:       tc.id,
					Name:     tc.name,
					AuthorID: tc.authorID,
				}, nil)
			}
			pr, err := usecase.PullRequestCreate(ctx, tc.id, tc.name, tc.authorID)

			if tc.errExpected {
				require.Empty(t, pr)
				require.Error(t, err)
			} else {
				require.NotEmpty(t, pr)
				require.NoError(t, err)
				require.Equal(t, tc.authorID, pr.AuthorId)
				require.Equal(t, tc.name, pr.PullRequestName)
				require.Equal(t, tc.id, pr.PullRequestId)
			}
		}
	})

	t.Run("pr merge", func(t *testing.T) {
		t.Parallel()
		tcs := []testCase{
			{
				id:          "valid",
				errExpected: false,
			},
			{
				id:          "invalid",
				errExpected: true,
			},
		}

		for _, tc := range tcs {
			if tc.errExpected {
				prRepo.EXPECT().MergePullRequest(ctx, tc.id).Return(entity.PR{}, errors.New("some error"))
			} else {
				prRepo.EXPECT().MergePullRequest(ctx, tc.id).Return(entity.PR{
					ID:       tc.id,
					MergedAt: time.Now(),
					Status:   entity.MERGED,
				}, nil)
			}
			pr, err := usecase.PullRequestMerge(ctx, tc.id)

			if tc.errExpected {
				require.Empty(t, pr)
				require.Error(t, err)
			} else {
				require.NotEmpty(t, pr)
				require.NoError(t, err)
				require.NotEmpty(t, pr.MergedAt)
				require.Equal(t, generated.PullRequestStatus(generated.PullRequestShortStatusMERGED), pr.Status)
				require.Equal(t, tc.id, pr.PullRequestId)
			}
		}
	})

	t.Run("pr reassign", func(t *testing.T) {
		t.Parallel()

		tcs := []testCase{
			{
				id:        "valid-open",
				oldUserID: "u1",
				status:    entity.OPEN,
				replaced:  "u2",
			},
			{
				id:        "valid-merged",
				oldUserID: "u1",
				status:    entity.MERGED,
				replaced:  "u3",
			},
			{
				id:          "invalid",
				oldUserID:   "u1",
				errExpected: true,
			},
		}

		for _, tc := range tcs {
			if tc.errExpected {
				prRepo.EXPECT().
					ReassignPullRequest(ctx, tc.id, tc.oldUserID).
					Return(entity.PR{}, "", errors.New("some error"))
			} else {
				prRepo.EXPECT().
					ReassignPullRequest(ctx, tc.id, tc.oldUserID).
					Return(entity.PR{
						ID:       tc.id,
						Status:   tc.status,
						AuthorID: "author",
					}, tc.replaced, nil)
			}

			pr, replaced, err := usecase.PullRequestReassign(ctx, tc.id, tc.oldUserID)

			if tc.errExpected {
				require.Error(t, err)
				require.Empty(t, pr)
				require.Empty(t, replaced)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, pr)
				require.Equal(t, tc.id, pr.PullRequestId)
				require.Equal(t, tc.replaced, replaced)

				expectedStatus := generated.PullRequestStatus(generated.PullRequestShortStatusOPEN)
				if tc.status == entity.MERGED {
					expectedStatus = generated.PullRequestStatus(generated.PullRequestShortStatusMERGED)
				}
				require.Equal(t, expectedStatus, pr.Status)
			}
		}
	})
}
