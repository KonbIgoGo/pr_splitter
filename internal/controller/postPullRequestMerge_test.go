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

func TestPRHandlers_Merge(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		body        any
		setupMock   func(m *mocks.MockPRUseCase)
		wantStatus  int
		wantErrCode *generated.ErrorResponseErrorCode
		wantPR      *generated.PullRequest
	}

	tests := []testCase{
		{
			name: "merge_success",
			body: generated.PostPullRequestMergeJSONBody{
				PullRequestId: "pr-1",
			},
			setupMock: func(m *mocks.MockPRUseCase) {
				m.EXPECT().
					PullRequestMerge(gomock.Any(), "pr-1").
					Return(generated.PullRequest{
						PullRequestId:   "pr-1",
						PullRequestName: "Feature X",
						AuthorId:        "u1",
						Status:          generated.PullRequestStatusMERGED,
						AssignedReviewers: []string{
							"u2", "u3",
						},
					}, nil)
			},
			wantStatus: http.StatusOK,
			wantPR: &generated.PullRequest{
				PullRequestId:   "pr-1",
				PullRequestName: "Feature X",
				AuthorId:        "u1",
				Status:          generated.PullRequestStatusMERGED,
				AssignedReviewers: []string{
					"u2", "u3",
				},
			},
		},
		{
			name: "merge_invalid_body_should_bind_error",
			body: `{"pull_request_id":`,
			setupMock: func(m *mocks.MockPRUseCase) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "merge_not_found",
			body: generated.PostPullRequestMergeJSONBody{
				PullRequestId: "pr-not-found",
			},
			setupMock: func(m *mocks.MockPRUseCase) {
				m.EXPECT().
					PullRequestMerge(gomock.Any(), "pr-not-found").
					Return(generated.PullRequest{}, entity.ErrPRNotFound)
			},
			wantStatus:  http.StatusNotFound,
			wantErrCode: ptrErrorCode(generated.NOTFOUND),
		},
		{
			name: "merge_internal_error",
			body: generated.PostPullRequestMergeJSONBody{
				PullRequestId: "pr-err",
			},
			setupMock: func(m *mocks.MockPRUseCase) {
				m.EXPECT().
					PullRequestMerge(gomock.Any(), "pr-err").
					Return(generated.PullRequest{}, assertAnyError{})
			},
			wantStatus: http.StatusInternalServerError,
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

			req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			require.Equal(t, tc.wantStatus, w.Code)

			if tc.wantStatus == http.StatusBadRequest && tc.wantErrCode == nil && tc.wantPR == nil {
				return
			}

			if tc.wantStatus == http.StatusOK {
				var resp generated.PullRequest
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)

				require.Equal(t, tc.wantPR.PullRequestId, resp.PullRequestId)
				require.Equal(t, tc.wantPR.PullRequestName, resp.PullRequestName)
				require.Equal(t, tc.wantPR.AuthorId, resp.AuthorId)
				require.Equal(t, tc.wantPR.Status, resp.Status)
				require.Equal(t, tc.wantPR.AssignedReviewers, resp.AssignedReviewers)
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

type assertAnyError struct{}

func (assertAnyError) Error() string { return "some error" }
