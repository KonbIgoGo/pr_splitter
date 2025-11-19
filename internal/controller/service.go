package controller

import (
	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/KonbIgoGo/pr_splitter/internal/usecase"
	"go.uber.org/zap"
)

var _ generated.ServerInterface = (*implementation)(nil)

type implementation struct {
	logger *zap.Logger
}

func New(
	logger *zap.Logger,
	prUseCase usecase.PRUseCase,
	userUseCase usecase.UserUseCase,
	teamUseCase usecase.TeamUseCase,
) *implementation {
	return &implementation{
		logger: logger,
	}
}
