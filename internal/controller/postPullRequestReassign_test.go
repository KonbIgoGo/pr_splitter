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

func TestPRHandlers_Reassign(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		body        any
		setupMock   func(m *mocks.MockPRUseCase)
		wantStatus  int
		wantErrCode *generated.ErrorResponseErrorCode
		wantResp    *reassignPRREsp
	}

	tests := []testCase{
		{
			name: "reassign_success",
			body: generated.PostPullRequestReassignJSONRequestBody{
				PullRequestId: "pr-1",
				OldUserId:     "u-old",
			},
			setupMock: func(m *mocks.MockPRUseCase) {
				m.EXPECT().
					PullRequestReassign(gomock.Any(), "pr-1", "u-old").
					Return(
						generated.PullRequest{
							PullRequestId:     "pr-1",
							PullRequestName:   "Feature X",
							AuthorId:          "u-author",
							Status:            generated.PullRequestStatusOPEN,
							AssignedReviewers: []string{"u-new", "u-other"},
						},
						"u-new",
						nil,
					)
			},
			wantStatus: http.StatusOK,
			wantResp: &reassignPRREsp{
				Pr: generated.PullRequest{
					PullRequestId:     "pr-1",
					PullRequestName:   "Feature X",
					AuthorId:          "u-author",
					Status:            generated.PullRequestStatusOPEN,
					AssignedReviewers: []string{"u-new", "u-other"},
				},
				ReplacedBy: "u-new",
			},
		},
		{
			name:       "reassign_bind_error",
			body:       `{"pull_request_id":`,
			setupMock:  func(m *mocks.MockPRUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "reassign_not_found",
			body: generated.PostPullRequestReassignJSONRequestBody{
				PullRequestId: "pr-missing",
				OldUserId:     "u-old",
			},
			setupMock: func(m *mocks.MockPRUseCase) {
				m.EXPECT().
					PullRequestReassign(gomock.Any(), "pr-missing", "u-old").
					Return(generated.PullRequest{}, "", entity.ErrPRNotFound)
			},
			wantStatus:  http.StatusNotFound,
			wantErrCode: func() *generated.ErrorResponseErrorCode { c := generated.NOTFOUND; return &c }(),
		},
		{
			name: "reassign_conflict_merged",
			body: generated.PostPullRequestReassignJSONRequestBody{
				PullRequestId: "pr-merged",
				OldUserId:     "u-old",
			},
			setupMock: func(m *mocks.MockPRUseCase) {
				m.EXPECT().
					PullRequestReassign(gomock.Any(), "pr-merged", "u-old").
					Return(generated.PullRequest{}, "", entity.ErrPRMerged)
			},
			wantStatus:  http.StatusConflict,
			wantErrCode: func() *generated.ErrorResponseErrorCode { c := generated.PRMERGED; return &c }(),
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
				tc.setupMock(prUC)
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

			req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			require.Equal(t, tc.wantStatus, w.Code)

			if tc.wantStatus == http.StatusBadRequest && tc.wantErrCode == nil && tc.wantResp == nil {
				return
			}

			if tc.wantStatus == http.StatusOK {
				var resp reassignPRREsp
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)

				require.Equal(t, tc.wantResp.ReplacedBy, resp.ReplacedBy)
				require.Equal(t, tc.wantResp.Pr.PullRequestId, resp.Pr.PullRequestId)
				require.Equal(t, tc.wantResp.Pr.PullRequestName, resp.Pr.PullRequestName)
				require.Equal(t, tc.wantResp.Pr.AuthorId, resp.Pr.AuthorId)
				require.Equal(t, tc.wantResp.Pr.Status, resp.Pr.Status)
				require.Equal(t, tc.wantResp.Pr.AssignedReviewers, resp.Pr.AssignedReviewers)
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
