package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/KonbIgoGo/pr_splitter/internal/entity"
	"github.com/KonbIgoGo/pr_splitter/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestTeamUseCases(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	teamRepo := mocks.NewMockTeamRepository(ctrl)
	usecase := New(nil, nil, nil, teamRepo)

	t.Run("team add", func(t *testing.T) {
		t.Parallel()

		type testCase struct {
			name        string
			members     []generated.TeamMember
			errExpected bool
		}

		tcs := []testCase{
			{
				name: "team-valid",
				members: []generated.TeamMember{
					{
						UserId:   "u1",
						Username: "user1",
						IsActive: true,
					},
					{
						UserId:   "u2",
						Username: "user2",
						IsActive: false,
					},
				},
				errExpected: false,
			},
			{
				name: "team-error",
				members: []generated.TeamMember{
					{
						UserId:   "u3",
						Username: "user3",
						IsActive: true,
					},
				},
				errExpected: true,
			},
		}

		for _, tc := range tcs {
			expectedMembers := make([]entity.TeamMember, 0, len(tc.members))
			for _, m := range tc.members {
				expectedMembers = append(expectedMembers, entity.TeamMember{
					UserID:   m.UserId,
					Username: m.Username,
					IsActive: m.IsActive,
				})
			}

			if tc.errExpected {
				teamRepo.EXPECT().
					AddTeam(ctx, entity.Team{
						TeamName: tc.name,
						Members:  expectedMembers,
					}).
					Return(entity.Team{}, errors.New("some error"))
			} else {
				teamRepo.EXPECT().
					AddTeam(ctx, entity.Team{
						TeamName: tc.name,
						Members:  expectedMembers,
					}).
					Return(entity.Team{
						TeamName: tc.name,
						Members:  expectedMembers,
					}, nil)
			}

			team, err := usecase.TeamAdd(ctx, tc.name, tc.members)

			if tc.errExpected {
				require.Error(t, err)
				require.Empty(t, team)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, team)
				require.Equal(t, tc.name, team.TeamName)
				require.Len(t, team.Members, len(tc.members))
				require.Equal(t, tc.members, team.Members)
			}
		}
	})

	t.Run("team get", func(t *testing.T) {
		t.Parallel()

		type testCase struct {
			name        string
			members     []entity.TeamMember
			errExpected bool
		}

		tcs := []testCase{
			{
				name: "team-valid",
				members: []entity.TeamMember{
					{
						UserID:   "u1",
						Username: "user1",
						IsActive: true,
					},
					{
						UserID:   "u2",
						Username: "user2",
						IsActive: false,
					},
				},
				errExpected: false,
			},
			{
				name:        "team-not-found",
				members:     nil,
				errExpected: true,
			},
		}

		for _, tc := range tcs {
			if tc.errExpected {
				teamRepo.EXPECT().
					GetTeam(ctx, tc.name).
					Return(entity.Team{}, errors.New("some error"))
			} else {
				teamRepo.EXPECT().
					GetTeam(ctx, tc.name).
					Return(entity.Team{
						TeamName: tc.name,
						Members:  tc.members,
					}, nil)
			}

			team, err := usecase.TeamGet(ctx, tc.name)

			if tc.errExpected {
				require.Error(t, err)
				require.Empty(t, team)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, team)
				require.Equal(t, tc.name, team.TeamName)

				expectedMembers := parseTeamMembers(tc.members)
				require.Equal(t, expectedMembers, team.Members)
			}
		}
	})
}
