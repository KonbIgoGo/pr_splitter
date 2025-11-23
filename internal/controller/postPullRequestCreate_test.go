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

func TestPRHandlers_Create(t *testing.T) {
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
			name: "create_success",
			body: generated.PostPullRequestCreateJSONBody{
				PullRequestId:   "pr-1",
				PullRequestName: "Feature X",
				AuthorId:        "u1",
			},
			setupMock: func(m *mocks.MockPRUseCase) {
				m.EXPECT().
					PullRequestCreate(gomock.Any(), "pr-1", "Feature X", "u1").
					Return(generated.PullRequest{
						PullRequestId:   "pr-1",
						PullRequestName: "Feature X",
						AuthorId:        "u1",
						Status:          generated.PullRequestStatusOPEN,
						AssignedReviewers: []string{
							"u2", "u3",
						},
					}, nil)
			},
			wantStatus: http.StatusCreated,
			wantPR: &generated.PullRequest{
				PullRequestId:   "pr-1",
				PullRequestName: "Feature X",
				AuthorId:        "u1",
				Status:          generated.PullRequestStatusOPEN,
				AssignedReviewers: []string{
					"u2", "u3",
				},
			},
		},
		{
			name:       "create_invalid_body_should_bind_error",
			body:       `{"pull_request_id": "pr-2", "pull_request_name":`,
			setupMock:  nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "create_author_not_found",
			body: generated.PostPullRequestCreateJSONBody{
				PullRequestId:   "pr-404",
				PullRequestName: "Feature Y",
				AuthorId:        "unknown",
			},
			setupMock: func(m *mocks.MockPRUseCase) {
				m.EXPECT().
					PullRequestCreate(gomock.Any(), "pr-404", "Feature Y", "unknown").
					Return(generated.PullRequest{}, entity.ErrUserNotFound)
			},
			wantStatus:  http.StatusNotFound,
			wantErrCode: ptrErrorCode(generated.NOTFOUND),
		},
		{
			name: "create_pr_exists_conflict",
			body: generated.PostPullRequestCreateJSONBody{
				PullRequestId:   "pr-dup",
				PullRequestName: "Duplicate",
				AuthorId:        "u1",
			},
			setupMock: func(m *mocks.MockPRUseCase) {
				m.EXPECT().
					PullRequestCreate(gomock.Any(), "pr-dup", "Duplicate", "u1").
					Return(generated.PullRequest{}, entity.ErrPRAlreadyExists)
			},
			wantStatus:  http.StatusConflict,
			wantErrCode: ptrErrorCode(generated.PREXISTS),
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

			req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			require.Equal(t, tc.wantStatus, w.Code)

			if tc.wantStatus == http.StatusBadRequest && tc.wantErrCode == nil && tc.wantPR == nil {
				return
			}

			if tc.wantStatus == http.StatusCreated {
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
