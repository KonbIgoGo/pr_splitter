package controller

import (
	"bytes"
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

func TestTeamHandlers_AddTeam(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		body        any
		setupMock   func(m *mocks.MockTeamUseCase)
		wantStatus  int
		wantErrCode *generated.ErrorResponseErrorCode
		wantTeam    *generated.Team
	}

	members := []generated.TeamMember{
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
	}

	tests := []testCase{
		{
			name: "team_add_success",
			body: generated.Team{
				TeamName: "backend",
				Members:  members,
			},
			setupMock: func(m *mocks.MockTeamUseCase) {
				m.EXPECT().
					TeamAdd(gomock.Any(), "backend", members).
					Return(generated.Team{
						TeamName: "backend",
						Members:  members,
					}, nil)
			},
			wantStatus: http.StatusCreated,
			wantTeam: &generated.Team{
				TeamName: "backend",
				Members:  members,
			},
		},
		{
			name:       "team_add_bind_error",
			body:       `{"team_name":`,
			setupMock:  func(m *mocks.MockTeamUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "team_add_team_exists",
			body: generated.Team{
				TeamName: "backend",
				Members:  members,
			},
			setupMock: func(m *mocks.MockTeamUseCase) {
				m.EXPECT().
					TeamAdd(gomock.Any(), "backend", members).
					Return(generated.Team{}, entity.ErrTeamAlreadyExists)
			},
			wantStatus: http.StatusConflict,
			wantErrCode: func() *generated.ErrorResponseErrorCode {
				c := generated.TEAMEXISTS
				return &c
			}(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			prUC := mocks.NewMockPRUseCase(ctrl)
			userUC := mocks.NewMockUserUseCase(ctrl)
			teamUC := mocks.NewMockTeamUseCase(ctrl)

			if tc.setupMock != nil {
				tc.setupMock(teamUC)
			}

			impl := &implementation{
				prUseCase:   prUC,
				userUseCase: userUC,
				teamUseCase: teamUC,
			}

			router := gin.New()
			generated.RegisterHandlers(router, impl)

			var reqBody []byte
			switch b := tc.body.(type) {
			case string:
				reqBody = []byte(b)
			default:
				var err error
				reqBody, err = json.Marshal(b)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			require.Equal(t, tc.wantStatus, w.Code)

			if tc.wantStatus == http.StatusBadRequest && tc.wantErrCode == nil && tc.wantTeam == nil {
				return
			}

			if tc.wantStatus == http.StatusCreated {
				var resp generated.Team
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)

				require.Equal(t, tc.wantTeam.TeamName, resp.TeamName)
				require.Len(t, resp.Members, len(tc.wantTeam.Members))
				require.Equal(t, tc.wantTeam.Members, resp.Members)
				return
			}

			if tc.wantErrCode != nil {
				var errResp generated.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &errResp)
				require.NoError(t, err)
				require.Equal(t, *tc.wantErrCode, errResp.Error.Code)
			}
		})
	}
}
