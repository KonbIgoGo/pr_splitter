package controller

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	// один раз перед всеми тестами
	gin.SetMode(gin.TestMode)

	code := m.Run()
	os.Exit(code)
}
