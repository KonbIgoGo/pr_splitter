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

func TestUserHandlers_SetIsActive(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		body        any
		setupMock   func(m *mocks.MockUserUseCase)
		wantStatus  int
		wantUser    *generated.User
		wantErrCode *generated.ErrorResponseErrorCode
	}

	tcs := []testCase{
		{
			name: "user_set_is_active_success",
			body: generated.PostUsersSetIsActiveJSONBody{
				UserId:   "u1",
				IsActive: true,
			},
			setupMock: func(m *mocks.MockUserUseCase) {
				m.EXPECT().
					UserSetIsActive(gomock.Any(), "u1", true).
					Return(generated.User{
						UserId:   "u1",
						Username: "Alice",
						TeamName: "backend",
						IsActive: true,
					}, nil)
			},
			wantStatus: http.StatusOK,
			wantUser: &generated.User{
				UserId:   "u1",
				Username: "Alice",
				TeamName: "backend",
				IsActive: true,
			},
		},
		{
			name: "user_set_is_active_not_found",
			body: generated.PostUsersSetIsActiveJSONBody{
				UserId:   "u404",
				IsActive: false,
			},
			setupMock: func(m *mocks.MockUserUseCase) {
				m.EXPECT().
					UserSetIsActive(gomock.Any(), "u404", false).
					Return(generated.User{}, entity.ErrUserNotFound)
			},
			wantStatus:  http.StatusNotFound,
			wantErrCode: func() *generated.ErrorResponseErrorCode { c := generated.NOTFOUND; return &c }(),
		},
		{
			name:       "user_set_is_active_bad_request_bind_error",
			body:       `{"user_id": "u1", "is_active": "not_bool"}`,
			setupMock:  func(m *mocks.MockUserUseCase) {},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserUC := mocks.NewMockUserUseCase(ctrl)

			if tc.setupMock != nil {
				tc.setupMock(mockUserUC)
			}

			impl := &implementation{
				userUseCase: mockUserUC,
			}

			router := gin.New()
			router.POST("/users/setIsActive", impl.PostUsersSetIsActive)

			var reqBody []byte
			var err error

			switch v := tc.body.(type) {
			case string:
				reqBody = []byte(v)
			default:
				reqBody, err = json.Marshal(v)
				require.NoError(t, err)
			}

			req, err := http.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			require.Equal(t, tc.wantStatus, rr.Code)

			if tc.wantUser != nil {
				var got generated.User
				err = json.Unmarshal(rr.Body.Bytes(), &got)
				require.NoError(t, err)
				require.Equal(t, *tc.wantUser, got)
			}

			if tc.wantErrCode != nil {
				var errResp generated.ErrorResponse
				err = json.Unmarshal(rr.Body.Bytes(), &errResp)
				require.NoError(t, err)
				require.Equal(t, *tc.wantErrCode, errResp.Error.Code)
			}
		})
	}
}
