package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (i *implementation) GetUsersGetPRAuthorityStatistic(c *gin.Context) {
	stat, err := i.userUseCase.UsersGetPRAuthorityStatistic(c.Request.Context())
	if err != nil {
		code, errRes := convertErrors(err)
		c.JSON(code, errRes)
		return
	}
	c.JSON(http.StatusOK, stat)
}
