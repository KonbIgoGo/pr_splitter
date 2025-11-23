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

func ptrErrorCode(c generated.ErrorResponseErrorCode) *generated.ErrorResponseErrorCode {
	return &c
}

func TestUserHandlers(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		userID      string
		setupMock   func(m *mocks.MockUserUseCase)
		wantStatus  int
		wantErrCode *generated.ErrorResponseErrorCode
		wantResp    *userGetReviewResp
	}

	tests := []testCase{
		{
			name:   "user_getReview_success",
			userID: "u1",
			setupMock: func(m *mocks.MockUserUseCase) {
				prs := []generated.PullRequestShort{
					{
						PullRequestId:   "pr-1",
						PullRequestName: "Add feature",
						AuthorId:        "author-1",
						Status:          generated.PullRequestShortStatusOPEN,
					},
					{
						PullRequestId:   "pr-2",
						PullRequestName: "Fix bug",
						AuthorId:        "author-2",
						Status:          generated.PullRequestShortStatusMERGED,
					},
				}

				m.EXPECT().
					UserGetReview(gomock.Any(), "u1").
					Return(prs, nil)
			},
			wantStatus: http.StatusOK,
			wantResp: &userGetReviewResp{
				UserID: "u1",
				PullRequests: []generated.PullRequestShort{
					{
						PullRequestId:   "pr-1",
						PullRequestName: "Add feature",
						AuthorId:        "author-1",
						Status:          generated.PullRequestShortStatusOPEN,
					},
					{
						PullRequestId:   "pr-2",
						PullRequestName: "Fix bug",
						AuthorId:        "author-2",
						Status:          generated.PullRequestShortStatusMERGED,
					},
				},
			},
		},
		{
			name:   "user_getReview_not_found",
			userID: "unknown",
			setupMock: func(m *mocks.MockUserUseCase) {
				m.EXPECT().
					UserGetReview(gomock.Any(), "unknown").
					Return(nil, entity.ErrUserNotFound)
			},
			wantStatus:  http.StatusNotFound,
			wantErrCode: ptrErrorCode(generated.NOTFOUND),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(func() { ctrl.Finish() })

			prUC := mocks.NewMockPRUseCase(ctrl)
			userUC := mocks.NewMockUserUseCase(ctrl)
			teamUC := mocks.NewMockTeamUseCase(ctrl)

			if tc.setupMock != nil {
				tc.setupMock(userUC)
			}

			impl := &implementation{
				prUseCase:   prUC,
				userUseCase: userUC,
				teamUseCase: teamUC,
			}

			router := gin.New()
			generated.RegisterHandlers(router, impl)

			req := httptest.NewRequest(
				http.MethodGet,
				"/users/getReview?user_id="+tc.userID,
				nil,
			)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			require.Equal(t, tc.wantStatus, w.Code)

			if tc.wantStatus == http.StatusOK {
				var resp userGetReviewResp
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)

				require.Equal(t, tc.wantResp.UserID, resp.UserID)
				require.Len(t, resp.PullRequests, len(tc.wantResp.PullRequests))

				for i := range tc.wantResp.PullRequests {
					require.Equal(t, tc.wantResp.PullRequests[i].PullRequestId, resp.PullRequests[i].PullRequestId)
					require.Equal(t, tc.wantResp.PullRequests[i].PullRequestName, resp.PullRequests[i].PullRequestName)
					require.Equal(t, tc.wantResp.PullRequests[i].AuthorId, resp.PullRequests[i].AuthorId)
					require.Equal(t, tc.wantResp.PullRequests[i].Status, resp.PullRequests[i].Status)
				}
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
