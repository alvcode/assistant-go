package clicontroller

import (
	"assistant-go/internal/config"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/spf13/cobra"
)

func InitCliCommands(rootCmd *cobra.Command, ctx context.Context, cfg *config.Config, db *pgxpool.Pool, minio *minio.Client) {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "clean-db",
		Short: "Cleaning the database from obsolete records",
		Run: func(cmd *cobra.Command, args []string) {
			CleanDB(ctx, cfg, db, minio)
		}})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "user-reset-password <login> <password>",
		Short: "Reset the password for a user",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			login := args[0]
			password := args[1]
			UserResetPassword(ctx, cfg, db, minio, login, password)
		}})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "user-register <login> <password>",
		Short: "Register a user",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			login := args[0]
			password := args[1]
			UserRegister(ctx, cfg, db, minio, login, password)
		}})
}
