package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/KonbIgoGo/pr_splitter/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserHandlers_GetReviewStatistic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUseCase(ctrl)

	impl := &implementation{
		userUseCase: userUC,
	}

	router := gin.New()
	router.GET("/users/getReviewStatistic", func(c *gin.Context) {
		impl.GetUsersGetReviewStatistic(c)
	})

	t.Run("success", func(t *testing.T) {
		expected := []generated.UserPRAuthorityStatistic{
			{
				UserId:   "u1",
				PrsCount: 2,
			},
			{
				UserId:   "u2",
				PrsCount: 5,
			},
		}

		userUC.EXPECT().
			UsersGetReviewStatistic(gomock.Any()).
			Return(expected, nil)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users/getReviewStatistic", nil)

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var got []generated.UserPRAuthorityStatistic
		err := json.Unmarshal(w.Body.Bytes(), &got)
		require.NoError(t, err)
		require.Equal(t, expected, got)
	})

	t.Run("internal_error", func(t *testing.T) {
		userUC.EXPECT().
			UsersGetReviewStatistic(gomock.Any()).
			Return(nil, errors.New("some error"))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users/getReviewStatistic", nil)

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)

		var errResp generated.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		require.NoError(t, err)
		require.Equal(t, "some error", errResp.Error.Message)
	})
}
