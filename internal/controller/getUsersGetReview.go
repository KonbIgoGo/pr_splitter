package controller

import (
	"net/http"

	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/gin-gonic/gin"
)

type userGetReviewResp struct {
	UserID       string                       `json:"user_id"`
	PullRequests []generated.PullRequestShort `json:"pull_requests"`
}

func (i *implementation) GetUsersGetReview(c *gin.Context, params generated.GetUsersGetReviewParams) {
	id := params.UserId
	prs, err := i.userUseCase.UserGetReview(c.Request.Context(), id)
	if err != nil {
		code, errRes := convertErrors(err)
		c.JSON(code, errRes)
		return
	}

	c.JSON(http.StatusOK, userGetReviewResp{id, prs})
}
