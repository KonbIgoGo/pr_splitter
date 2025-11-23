package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/KonbIgoGo/pr_splitter/internal/entity"
	"github.com/KonbIgoGo/pr_splitter/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestTeamHandlers(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(func() { ctrl.Finish() })

	teamUC := mocks.NewMockTeamUseCase(ctrl)

	impl := &implementation{
		teamUseCase: teamUC,
	}

	type testCase struct {
		name        string
		teamName    string
		team        generated.Team
		err         error
		errExpected bool
	}

	t.Run("team get", func(t *testing.T) {
		t.Parallel()

		tcs := []testCase{
			{
				name:     "success",
				teamName: "backend",
				team: generated.Team{
					TeamName: "backend",
					Members: []generated.TeamMember{
						{
							UserId:   "u1",
							Username: "Alice",
							IsActive: true,
						},
						{
							UserId:   "u2",
							Username: "Bob",
							IsActive: true,
						},
					},
				},
				errExpected: false,
			},
			{
				name:        "team not found",
				teamName:    "missing",
				err:         entity.ErrTeamNotFound,
				errExpected: true,
			},
		}

		for _, tc := range tcs {
			if tc.errExpected {
				teamUC.EXPECT().
					TeamGet(gomock.Any(), tc.teamName).
					Return(generated.Team{}, tc.err)
			} else {
				teamUC.EXPECT().
					TeamGet(gomock.Any(), tc.teamName).
					Return(tc.team, nil)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, err := http.NewRequest(
				http.MethodGet,
				"/team/get?team_name="+tc.teamName,
				nil,
			)
			require.NoError(t, err)
			c.Request = req

			params := generated.GetTeamGetParams{
				TeamName: tc.teamName,
			}

			impl.GetTeamGet(c, params)

			if tc.errExpected {
				require.Equal(t, http.StatusNotFound, w.Code)

				var resp generated.ErrorResponse
				err = json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				require.Equal(t, generated.NOTFOUND, resp.Error.Code)
			} else {
				require.Equal(t, http.StatusOK, w.Code)

				var got generated.Team
				err = json.Unmarshal(w.Body.Bytes(), &got)
				require.NoError(t, err)

				require.Equal(t, tc.team.TeamName, got.TeamName)
				require.Equal(t, tc.team.Members, got.Members)
			}
		}
	})
}
