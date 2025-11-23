package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KonbIgoGo/pr_splitter/config"
	"github.com/KonbIgoGo/pr_splitter/db"
	"github.com/KonbIgoGo/pr_splitter/generated"
	"github.com/KonbIgoGo/pr_splitter/internal/controller"
	"github.com/KonbIgoGo/pr_splitter/internal/repository"
	"github.com/KonbIgoGo/pr_splitter/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func Run(logger *zap.Logger, cfg *config.Config) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dbPool, err := pgxpool.New(ctx, cfg.PG.URL)

	if err != nil {
		logger.Error("can not create pgxpool", zap.Error(err))
		os.Exit(1)
	}
	defer dbPool.Close()

	db.SetupPostgres(dbPool, logger)

	repo := repository.NewPostgresRepository(dbPool)

	useCases := usecase.New(logger, repo, repo, repo)
	ctrl := controller.New(logger, useCases, useCases, useCases)

	go runRest(cfg, ctrl)

	<-ctx.Done()
	time.Sleep(time.Second)

}

func runRest(cfg *config.Config, service generated.ServerInterface) {
	r := gin.Default()
	generated.RegisterHandlers(r, service)
	r.Run(":" + cfg.HTTP.Port)
}
