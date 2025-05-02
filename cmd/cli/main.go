package main

import (
	"assistant-go/cmd/cli/clicontroller"
	"assistant-go/internal/config"
	"assistant-go/internal/logging"
	"assistant-go/internal/storage/postgres"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var cfg *config.Config

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg = config.MustLoad()
	logger := logging.NewLogger(cfg.Env)
	ctx = logging.ContextWithLogger(ctx, logger)

	pgConfig := postgres.NewPgConfig(cfg.DB.Host, cfg.DB.Port, cfg.DB.Username, cfg.DB.Password, cfg.DB.Database)
	pgClient, err := postgres.NewClient(ctx, 5, time.Second*5, pgConfig)
	if err != nil {
		logging.GetLogger(ctx).Fatalln(err)
	}

	var minioClient *minio.Client
	if cfg.File.UploadPlace == config.FileUploadS3Place && cfg.S3.SecretAccessKey != "" {
		minioClient, err = minio.New(cfg.S3.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.S3.AccessKey, cfg.S3.SecretAccessKey, ""),
			Secure: cfg.S3.UseSSL,
		})
		if err != nil {
			logging.GetLogger(ctx).Fatalln(err)
		}
	}

	rootCmd := &cobra.Command{
		Use:   "ast",
		Short: "CLI commands",
	}

	// Добавляем команды
	clicontroller.InitCliCommands(rootCmd, ctx, cfg, pgClient, minioClient)

	// Запуск приложения
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
