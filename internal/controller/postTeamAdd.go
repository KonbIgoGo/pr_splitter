package controller

import (
	"net/http"

	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/gin-gonic/gin"
)

func (i *implementation) PostTeamAdd(c *gin.Context) {
	var body generated.PostTeamAddJSONRequestBody

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, struct{}{})
		return
	}

	team, err := i.teamUseCase.TeamAdd(c.Request.Context(), body.TeamName, body.Members)
	if err != nil {
		code, errRes := convertErrors(err)
		c.JSON(code, errRes)
		return
	}

	c.JSON(http.StatusCreated, team)
}
