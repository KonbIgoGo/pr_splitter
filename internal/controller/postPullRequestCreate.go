package controller

import (
	"net/http"

	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/gin-gonic/gin"
)

func (i *implementation) PostPullRequestCreate(c *gin.Context) {
	var body generated.PostPullRequestCreateJSONBody

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, struct{}{})
		return
	}

	pr, err := i.prUseCase.PullRequestCreate(c.Request.Context(), body.PullRequestId, body.PullRequestName, body.AuthorId)
	if err != nil {
		code, errRes := convertErrors(err)
		c.JSON(code, errRes)
		return
	}

	c.JSON(http.StatusCreated, pr)
}
