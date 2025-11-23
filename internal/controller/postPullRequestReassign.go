package controller

import (
	"net/http"

	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/gin-gonic/gin"
)

type reassignPRREsp struct {
	Pr         generated.PullRequest `json:"pr"`
	ReplacedBy string                `json:"replaced_by"`
}

func (i *implementation) PostPullRequestReassign(c *gin.Context) {
	var body generated.PostPullRequestReassignJSONRequestBody

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, struct{}{})
		return
	}

	pr, replaced, err := i.prUseCase.PullRequestReassign(c.Request.Context(), body.PullRequestId, body.OldUserId)
	if err != nil {
		code, errRes := convertErrors(err)
		c.JSON(code, errRes)
		return
	}

	c.JSON(http.StatusOK, reassignPRREsp{
		Pr:         pr,
		ReplacedBy: replaced,
	})
}
