package clicontroller

import (
	"assistant-go/internal/config"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
)

func InitCliCommands(rootCmd *cobra.Command, ctx context.Context, cfg *config.Config, db *pgxpool.Pool) {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "clean-db",
		Short: "Cleaning the database from obsolete records",
		Run: func(cmd *cobra.Command, args []string) {
			CleanDBInit(ctx, cfg, db)
		}})
}
