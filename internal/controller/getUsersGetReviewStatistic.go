package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (i *implementation) GetUsersGetReviewStatistic(c *gin.Context) {
	stat, err := i.userUseCase.UsersGetReviewStatistic(c.Request.Context())
	if err != nil {
		code, errRes := convertErrors(err)
		c.JSON(code, errRes)
		return
	}
	c.JSON(http.StatusOK, stat)
}
