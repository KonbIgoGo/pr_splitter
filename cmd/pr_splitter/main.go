package main

import (
	"github.com/KonbIgoGo/pr_splitter/config"
	"github.com/KonbIgoGo/pr_splitter/internal/app"
	log "github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

func main() {
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
