package controller

import (
	"net/http"

	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/gin-gonic/gin"
)

func (i *implementation) PostPullRequestMerge(c *gin.Context) {
	var body generated.PostPullRequestMergeJSONBody

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, struct{}{})
		return
	}

	pr, err := i.prUseCase.PullRequestMerge(c.Request.Context(), body.PullRequestId)
	if err != nil {
		code, errRes := convertErrors(err)
		c.JSON(code, errRes)
		return
	}

	c.JSON(http.StatusOK, pr)
}
