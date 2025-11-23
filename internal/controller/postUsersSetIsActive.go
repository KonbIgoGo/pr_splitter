package controller

import (
	"net/http"

	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/gin-gonic/gin"
)

func (i *implementation) PostUsersSetIsActive(c *gin.Context) {
	var body generated.PostUsersSetIsActiveJSONBody

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusConflict, struct{}{})
		return
	}

	user, err := i.userUseCase.UserSetIsActive(c.Request.Context(), body.UserId, body.IsActive)
	if err != nil {
		code, errRes := convertErrors(err)
		c.JSON(code, errRes)
		return
	}

	c.JSON(http.StatusOK, user)
}
