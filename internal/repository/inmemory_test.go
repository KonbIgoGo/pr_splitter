package repository

import (
	"context"
	"testing"

	"github.com/KonbIgoGo/pr_splitter/internal/entity"
	"github.com/stretchr/testify/require"
)

func seedDefaultTeams(ctx context.Context, t *testing.T, repo *inmemoryImpl) {
	t.Helper()

	_, err := repo.AddTeam(ctx, entity.Team{
		Members: []entity.TeamMember{
			{
				UserID:   "u1",
				Username: "u1",
				IsActive: true,
			},
			{
				UserID:   "u2",
				Username: "u2",
				IsActive: true,
			},
			{
				UserID:   "u3",
				Username: "u3",
				IsActive: true,
			},
		},
		TeamName: "valid3",
	})
	require.NoError(t, err)

	_, err = repo.AddTeam(ctx, entity.Team{
		Members: []entity.TeamMember{
			{
				UserID:   "u1_2",
				Username: "u1_2",
				IsActive: true,
			},
			{
				UserID:   "u2_2",
				Username: "u2_2",
				IsActive: true,
			},
		},
		TeamName: "valid2",
	})
	require.NoError(t, err)

	_, err = repo.AddTeam(ctx, entity.Team{
		Members: []entity.TeamMember{
			{
				UserID:   "u1_3",
				Username: "u1_3",
				IsActive: true,
			},
		},
		TeamName: "valid1",
	})
	require.NoError(t, err)
}

func TestInmemory_AddTeam(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	type testCase struct {
		name      string
		seed      []entity.Team
		toAdd     entity.Team
		wantErr   error
		wantEmpty bool
	}

	tests := []testCase{
		{
			name: "success_empty_members",
			toAdd: entity.Team{
				TeamName: "team1",
				Members:  nil,
			},
			wantErr:   nil,
			wantEmpty: false,
		},
		{
			name: "success_with_members",
			toAdd: entity.Team{
				TeamName: "team2",
				Members: []entity.TeamMember{
					{UserID: "u1", Username: "u1", IsActive: true},
					{UserID: "u2", Username: "u2", IsActive: false},
				},
			},
			wantErr:   nil,
			wantEmpty: false,
		},
		{
			name: "duplicate_name",
			seed: []entity.Team{
				{
					TeamName: "dup",
					Members:  nil,
				},
			},
			toAdd: entity.Team{
				TeamName: "dup",
				Members:  nil,
			},
			wantErr:   entity.ErrTeamAlreadyExists,
			wantEmpty: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInmemoryRepository()

			for _, seed := range tc.seed {
				_, err := repo.AddTeam(ctx, seed)
				require.NoError(t, err)
			}

			got, err := repo.AddTeam(ctx, tc.toAdd)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Empty(t, got)
				return
			}

			require.NoError(t, err)
			require.False(t, tc.wantEmpty)
			require.Equal(t, tc.toAdd.TeamName, got.TeamName)

			stored, ok := repo.teamRepo[tc.toAdd.TeamName]
			require.True(t, ok)
			require.Equal(t, got, *stored)
		})
	}
}

func TestInmemory_GetTeam(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	type testCase struct {
		name      string
		seed      []entity.Team
		teamName  string
		wantErr   error
		wantEmpty bool
	}

	tests := []testCase{
		{
			name:      "not_found_without_seed",
			seed:      nil,
			teamName:  "no_such",
			wantErr:   entity.ErrTeamNotFound,
			wantEmpty: true,
		},
		{
			name: "found_after_add",
			seed: []entity.Team{
				{
					TeamName: "team1",
					Members: []entity.TeamMember{
						{UserID: "u1", Username: "u1", IsActive: true},
					},
				},
			},
			teamName:  "team1",
			wantErr:   nil,
			wantEmpty: false,
		},
		{
			name: "not_found_with_other_team",
			seed: []entity.Team{
				{
					TeamName: "team2",
					Members:  nil,
				},
			},
			teamName:  "absent",
			wantErr:   entity.ErrTeamNotFound,
			wantEmpty: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInmemoryRepository()

			for _, seed := range tc.seed {
				_, err := repo.AddTeam(ctx, seed)
				require.NoError(t, err)
			}

			got, err := repo.GetTeam(ctx, tc.teamName)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Empty(t, got)
				return
			}

			require.NoError(t, err)
			require.False(t, tc.wantEmpty)
			require.Equal(t, tc.teamName, got.TeamName)
		})
	}
}

func TestInmemory_CreatePullRequest(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name              string
		seed              func(ctx context.Context, t *testing.T, repo *inmemoryImpl)
		id                string
		prName            string
		authorID          string
		wantErr           error
		wantReviewersLen  int
		checkReviewersLen bool
	}

	emptySeed := func(_ context.Context, _ *testing.T, _ *inmemoryImpl) {}

	tests := []testCase{
		{
			name:              "three_members_two_reviewers",
			seed:              seedDefaultTeams,
			id:                "valid_create",
			prName:            "valid_create",
			authorID:          "u1",
			wantReviewersLen:  2,
			checkReviewersLen: true,
		},
		{
			name:              "one_member_zero_reviewers",
			seed:              seedDefaultTeams,
			id:                "empty_reviewers_create",
			prName:            "empty_reviewers_create",
			authorID:          "u1_3",
			wantReviewersLen:  0,
			checkReviewersLen: true,
		},
		{
			name:              "two_members_one_reviewer",
			seed:              seedDefaultTeams,
			id:                "one_reviewers_create",
			prName:            "one_reviewers_create",
			authorID:          "u1_2",
			wantReviewersLen:  1,
			checkReviewersLen: true,
		},
		{
			name: "duplicate_id",
			seed: func(ctx context.Context, t *testing.T, repo *inmemoryImpl) {
				t.Helper()
				seedDefaultTeams(ctx, t, repo)
				_, err := repo.CreatePullRequest(ctx, "dup", "dup", "u1")
				require.NoError(t, err)
			},
			id:                "dup",
			prName:            "dup",
			authorID:          "u1",
			wantErr:           entity.ErrPRAlreadyExists,
			checkReviewersLen: false,
		},
		{
			name:              "user_not_found",
			seed:              emptySeed,
			id:                "no_user",
			prName:            "no_user",
			authorID:          "unknown",
			wantErr:           entity.ErrUserNotFound,
			checkReviewersLen: false,
		},
		{
			name: "team_not_found",
			seed: func(_ context.Context, t *testing.T, repo *inmemoryImpl) {
				t.Helper()
				repo.userRepo["ghost"] = &entity.User{
					ID:       "ghost",
					Name:     "ghost",
					IsActive: true,
					TeamName: "missing_team",
				}
			},
			id:                "no_team",
			prName:            "no_team",
			authorID:          "ghost",
			wantErr:           entity.ErrTeamNotFound,
			checkReviewersLen: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			repo := NewInmemoryRepository()

			if tc.seed != nil {
				tc.seed(ctx, t, repo)
			}

			pr, err := repo.CreatePullRequest(ctx, tc.id, tc.prName, tc.authorID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Empty(t, pr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.id, pr.ID)
			require.Equal(t, tc.prName, pr.Name)
			require.Equal(t, tc.authorID, pr.AuthorID)
			require.Equal(t, entity.OPEN, pr.Status)
			require.False(t, pr.CreatedAt.IsZero())

			stored, ok := repo.prRepo[tc.id]
			require.True(t, ok)
			require.Equal(t, pr, *stored)

			if tc.checkReviewersLen {
				require.Len(t, pr.AssignedReviewersID, tc.wantReviewersLen)
			}

			if len(pr.AssignedReviewersID) > 0 {
				teamIDs, err := repo.getUserTeamIDs(tc.authorID)
				require.NoError(t, err)

				seen := make(map[string]struct{})
				for _, rID := range pr.AssignedReviewersID {
					require.NotEmpty(t, rID)
					require.NotEqual(t, tc.authorID, rID)

					_, already := seen[rID]
					require.False(t, already, "duplicate reviewer %s", rID)
					seen[rID] = struct{}{}

					require.Contains(t, teamIDs, rID)
				}
			}
		})
	}
}

func TestInmemory_MergePullRequest(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		seed    func(t *testing.T, repo *inmemoryImpl, ctx context.Context) string
		wantErr error
	}

	tests := []testCase{
		{
			name: "pr_not_found",
			seed: func(_ *testing.T, _ *inmemoryImpl, _ context.Context) string {
				return "unknown"
			},
			wantErr: entity.ErrPRNotFound,
		},
		{
			name: "merge_and_idempotent",
			seed: func(t *testing.T, repo *inmemoryImpl, ctx context.Context) string {
				t.Helper()
				seedDefaultTeams(ctx, t, repo)
				pr, err := repo.CreatePullRequest(ctx, "merge_id", "merge_name", "u1")
				require.NoError(t, err)
				return pr.ID
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			repo := NewInmemoryRepository()

			prID := tc.seed(t, repo, ctx)

			pr, err := repo.MergePullRequest(ctx, prID)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Empty(t, pr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, prID, pr.ID)
			require.Equal(t, entity.MERGED, pr.Status)
			require.False(t, pr.MergedAt.IsZero())

			pr2, err := repo.MergePullRequest(ctx, prID)
			require.NoError(t, err)
			require.Equal(t, entity.MERGED, pr2.Status)
			require.False(t, pr2.MergedAt.IsZero())
		})
	}
}

func TestInmemory_ReassignPullRequest(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		prepare func(t *testing.T, repo *inmemoryImpl, ctx context.Context) (string, string)
		wantErr error
	}

	tests := []testCase{
		{
			name: "success_reassign",
			prepare: func(t *testing.T, repo *inmemoryImpl, ctx context.Context) (string, string) {
				t.Helper()
				seedDefaultTeams(ctx, t, repo)
				pr, err := repo.CreatePullRequest(ctx, "reassign_id", "reassign", "u1")
				require.NoError(t, err)
				require.Len(t, pr.AssignedReviewersID, 2)
				return pr.ID, pr.AssignedReviewersID[0]
			},
			wantErr: nil,
		},
		{
			name: "pr_not_found",
			prepare: func(t *testing.T, repo *inmemoryImpl, ctx context.Context) (string, string) {
				t.Helper()
				seedDefaultTeams(ctx, t, repo)
				return "unknown_pr", "u1"
			},
			wantErr: entity.ErrPRNotFound,
		},
		{
			name: "old_user_not_reviewer",
			prepare: func(t *testing.T, repo *inmemoryImpl, ctx context.Context) (string, string) {
				t.Helper()
				seedDefaultTeams(ctx, t, repo)
				pr, err := repo.CreatePullRequest(ctx, "reassign2", "reassign2", "u1")
				require.NoError(t, err)
				return pr.ID, pr.AuthorID
			},
			wantErr: entity.ErrReviewerIsNotAssignedToPR,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			repo := NewInmemoryRepository()

			prID, oldUserID := tc.prepare(t, repo, ctx)

			got, _, err := repo.ReassignPullRequest(ctx, prID, oldUserID)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Empty(t, got)
				return
			}

			require.NoError(t, err)
			require.Equal(t, prID, got.ID)

			require.NotContains(t, got.AssignedReviewersID, oldUserID)

			teamIDs, err := repo.getUserTeamIDs(got.AuthorID)
			require.NoError(t, err)

			for _, rID := range got.AssignedReviewersID {
				require.NotEmpty(t, rID)
				require.NotEqual(t, got.AuthorID, rID)
				require.Contains(t, teamIDs, rID)
			}
		})
	}
}

func TestInmemory_SetIsActiveUser(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	type testCase struct {
		name      string
		seedTeams bool
		userID    string
		isActive  bool
		wantErr   error
	}

	tests := []testCase{
		{
			name:      "existing_user_set_false",
			seedTeams: true,
			userID:    "u1",
			isActive:  false,
			wantErr:   nil,
		},
		{
			name:      "existing_user_set_true",
			seedTeams: true,
			userID:    "u2",
			isActive:  true,
			wantErr:   nil,
		},
		{
			name:      "unknown_user",
			seedTeams: false,
			userID:    "unknown",
			isActive:  true,
			wantErr:   entity.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := NewInmemoryRepository()

			if tc.seedTeams {
				_, err := repo.AddTeam(ctx, entity.Team{
					TeamName: "team1",
					Members: []entity.TeamMember{
						{UserID: "u1", Username: "u1", IsActive: true},
						{UserID: "u2", Username: "u2", IsActive: false},
					},
				})
				require.NoError(t, err)
			}

			user, err := repo.SetIsActiveUser(ctx, tc.userID, tc.isActive)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Empty(t, user)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.userID, user.ID)
			require.Equal(t, tc.isActive, user.IsActive)
		})
	}
}

func TestInmemory_GetReviewUser(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	repo := NewInmemoryRepository()
	seedDefaultTeams(ctx, t, repo)

	pr1, err := repo.CreatePullRequest(ctx, "pr1", "pr1", "u1")
	require.NoError(t, err)

	pr2, err := repo.CreatePullRequest(ctx, "pr2", "pr2", "u2")
	require.NoError(t, err)

	created := []entity.PR{pr1, pr2}

	for _, pr := range created {
		for _, reviewerID := range pr.AssignedReviewersID {
			if reviewerID == "" {
				continue
			}

			reviewerID := reviewerID
			prID := pr.ID

			t.Run("reviewer_"+reviewerID+"_for_"+prID, func(t *testing.T) {
				t.Parallel()

				prs, err := repo.GetReviewUser(ctx, reviewerID)
				require.NoError(t, err)

				found := false
				for _, p := range prs {
					if p.ID == prID {
						found = true
						break
					}
				}
				require.True(t, found, "PR %s must be reviewed by %s", prID, reviewerID)
			})
		}
	}

	t.Run("unknown_user_no_prs", func(t *testing.T) {
		t.Parallel()

		prs, err := repo.GetReviewUser(ctx, "unknown_user")
		require.NoError(t, err)
		require.Empty(t, prs)
	})
}
