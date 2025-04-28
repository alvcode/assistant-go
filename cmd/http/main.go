package main

import (
	"assistant-go/internal/app"
	"assistant-go/internal/config"
	"assistant-go/internal/locale"
	"assistant-go/internal/logging"
	"assistant-go/pkg/vld"
	"context"
	"fmt"
	"os"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.MustLoad()

	logger := logging.NewLogger(cfg.Env)
	ctx = logging.ContextWithLogger(ctx, logger)

	logging.GetLogger(ctx).Infoln("Starting application")

	if _, err := os.Stat(cfg.File.SavePath); os.IsNotExist(err) {
		err := os.Mkdir(cfg.File.SavePath, 0755)
		if err != nil {
			logging.GetLogger(ctx).Fatalln(fmt.Errorf("could not create file save path: %s", err))
		}
	}

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

//func setupLogger(env string) *slog.Logger {
//	var log *slog.Logger
//
//	switch env {
//	case config.EnvDev:
//		log = slog.New(
//			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
//		)
//	case config.EnvTest:
//		log = slog.New(
//			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
//		)
//	case config.EnvProd:
//		log = slog.New(
//			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
//		)
//	}
//	return log
//}
