package controller

import (
	"net/http"

	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/gin-gonic/gin"
)

func (i *implementation) GetTeamGet(c *gin.Context, params generated.GetTeamGetParams) {
	name := params.TeamName
	team, err := i.teamUseCase.TeamGet(c.Request.Context(), name)
	if err != nil {
		status, errRes := convertErrors(err)
		c.JSON(status, errRes)
		return
	}

	c.JSON(http.StatusOK, team)
}
