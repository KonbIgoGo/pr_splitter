package usecase

import (
	"errors"
	"testing"

	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/KonbIgoGo/pr_splitter/internal/entity"
	"github.com/KonbIgoGo/pr_splitter/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserUseCases(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	ctrl := gomock.NewController(t)

	userRepo := mocks.NewMockUserRepository(ctrl)
	usecase := New(nil, nil, userRepo, nil)

	t.Run("user set is active", func(t *testing.T) {
		t.Parallel()

		type testCase struct {
			name        string
			id          string
			isActive    bool
			errExpected bool
		}

		tcs := []testCase{
			{
				name:        "set active success",
				id:          "u1",
				isActive:    true,
				errExpected: false,
			},
			{
				name:        "set inactive success",
				id:          "u2",
				isActive:    false,
				errExpected: false,
			},
			{
				name:        "repository error",
				id:          "u3",
				isActive:    true,
				errExpected: true,
			},
		}

		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				if tc.errExpected {
					userRepo.EXPECT().
						SetIsActiveUser(ctx, tc.id, tc.isActive).
						Return(entity.User{}, errors.New("some error"))
				} else {
					userRepo.EXPECT().
						SetIsActiveUser(ctx, tc.id, tc.isActive).
						Return(entity.User{
							ID:       tc.id,
							Name:     "user-" + tc.id,
							IsActive: tc.isActive,
							TeamName: "team-" + tc.id,
						}, nil)
				}

				user, err := usecase.UserSetIsActive(ctx, tc.id, tc.isActive)

				if tc.errExpected {
					require.Error(t, err)
					require.Empty(t, user)
				} else {
					require.NoError(t, err)
					require.NotEmpty(t, user)
					require.Equal(t, tc.id, user.UserId)
					require.Equal(t, tc.isActive, user.IsActive)
					require.Equal(t, "user-"+tc.id, user.Username)
					require.Equal(t, "team-"+tc.id, user.TeamName)
				}
			})
		}
	})

	t.Run("user get review", func(t *testing.T) {
		t.Parallel()

		type testCase struct {
			name        string
			id          string
			prs         []entity.PR
			errExpected bool
		}

		tcs := []testCase{
			{
				name: "no prs",
				id:   "u1",
				prs:  []entity.PR{},
			},
			{
				name: "several prs with different statuses",
				id:   "u2",
				prs: []entity.PR{
					{
						ID:       "pr-1",
						Name:     "first",
						AuthorID: "author-1",
						Status:   entity.OPEN,
					},
					{
						ID:       "pr-2",
						Name:     "second",
						AuthorID: "author-2",
						Status:   entity.MERGED,
					},
				},
			},
			{
				name:        "repository error",
				id:          "u3",
				prs:         nil,
				errExpected: true,
			},
		}

		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				if tc.errExpected {
					userRepo.EXPECT().
						GetReviewUser(ctx, tc.id).
						Return(nil, errors.New("some error"))
				} else {
					userRepo.EXPECT().
						GetReviewUser(ctx, tc.id).
						Return(tc.prs, nil)
				}

				prs, err := usecase.UserGetReview(ctx, tc.id)

				if tc.errExpected {
					require.Error(t, err)
					require.Nil(t, prs)
				} else {
					require.NoError(t, err)
					require.Len(t, prs, len(tc.prs))

					for idx, src := range tc.prs {
						dst := prs[idx]
						require.Equal(t, src.ID, dst.PullRequestId)
						require.Equal(t, src.Name, dst.PullRequestName)
						require.Equal(t, src.AuthorID, dst.AuthorId)

						expectedStatus := generated.PullRequestShortStatusOPEN
						if src.Status == entity.MERGED {
							expectedStatus = generated.PullRequestShortStatusMERGED
						}
						require.Equal(t, expectedStatus, dst.Status)
					}
				}
			})
		}
	})

	t.Run("users get review statistic", func(t *testing.T) {
		t.Parallel()

		type testCase struct {
			name        string
			stats       []entity.Statistic
			errExpected bool
		}

		tcs := []testCase{
			{
				name:  "no stats",
				stats: []entity.Statistic{},
			},
			{
				name: "several stats",
				stats: []entity.Statistic{
					{UserID: "u1", Data: 2},
					{UserID: "u2", Data: 5},
				},
			},
			{
				name:        "repository error",
				stats:       nil,
				errExpected: true,
			},
		}

		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				if tc.errExpected {
					userRepo.EXPECT().
						GetUsersReviewStatistic(ctx).
						Return(nil, errors.New("some error"))
				} else {
					userRepo.EXPECT().
						GetUsersReviewStatistic(ctx).
						Return(tc.stats, nil)
				}

				res, err := usecase.UsersGetReviewStatistic(ctx)

				if tc.errExpected {
					require.Error(t, err)
					require.Nil(t, res)
				} else {
					require.NoError(t, err)
					require.Len(t, res, len(tc.stats))

					for i, s := range tc.stats {
						require.Equal(t, s.UserID, res[i].UserId)
						require.Equal(t, s.Data, res[i].PrsCount)
					}
				}
			})
		}
	})

	t.Run("users get PR authority statistic", func(t *testing.T) {
		t.Parallel()

		type testCase struct {
			name        string
			stats       []entity.Statistic
			errExpected bool
		}

		tcs := []testCase{
			{
				name:  "no stats",
				stats: []entity.Statistic{},
			},
			{
				name: "several stats",
				stats: []entity.Statistic{
					{UserID: "u10", Data: 1},
					{UserID: "u20", Data: 3},
				},
			},
			{
				name:        "repository error",
				stats:       nil,
				errExpected: true,
			},
		}

		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				if tc.errExpected {
					userRepo.EXPECT().
						GetUsersPRAuthorityStatistic(ctx).
						Return(nil, errors.New("some error"))
				} else {
					userRepo.EXPECT().
						GetUsersPRAuthorityStatistic(ctx).
						Return(tc.stats, nil)
				}

				res, err := usecase.UsersGetPRAuthorityStatistic(ctx)

				if tc.errExpected {
					require.Error(t, err)
					require.Nil(t, res)
				} else {
					require.NoError(t, err)
					require.Len(t, res, len(tc.stats))

					for i, s := range tc.stats {
						require.Equal(t, s.UserID, res[i].UserId)
						require.Equal(t, s.Data, res[i].PrsCount)
					}
				}
			})
		}
	})
}
