package main

import (
	"os"

	"github.com/KonbIgoGo/pr_splitter/config"
	"github.com/KonbIgoGo/pr_splitter/internal/app"
	log "github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

func main() {

	os.Setenv("HTTP_PORT", "8080")
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_PORT", "5433")
	os.Setenv("POSTGRES_DB", "pr_db")
	os.Setenv("POSTGRES_USER", "postgres")
	os.Setenv("POSTGRES_PASSWORD", "1234")
	os.Setenv("POSTGRES_MAX_CONN", "10")
	cfg, err := config.New()

	if err != nil {
		log.Fatalf("cannot get application config: %s", err.Error())
	}

	logger, err := zap.NewProduction()

	if err != nil {
		log.Fatalf("cannot init logger: %s", err.Error())
	}
	defer logger.Sync()

	app.Run(logger, cfg)
}
