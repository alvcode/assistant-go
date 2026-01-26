package main

import (
	"assistant-go/internal/app"
	"assistant-go/internal/config"
	"assistant-go/internal/locale"
	"assistant-go/internal/logging"
	"assistant-go/pkg/vld"
	"context"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.MustLoad()

	logger := logging.NewLogger(cfg.Env)
	ctx = logging.ContextWithLogger(ctx, logger)

	logging.GetLogger(ctx).Infoln("Starting application")

	locale.InitLocales(ctx)
	vld.InitValidator(ctx)

	a, err := app.NewApp(ctx, cfg)
	if err != nil {
		logging.GetLogger(ctx).Fatalln(err)
	}

	logging.GetLogger(ctx).Println("Before Run")
	if err = a.Run(ctx); err != nil {
		logging.GetLogger(ctx).Fatalln(err)
	}
}
