package app

import (
	"assistant-go/internal/config"
	"assistant-go/internal/controller"
	"assistant-go/internal/logging"
	"assistant-go/internal/storage/postgres"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/cors"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"strings"
	"time"

	_ "assistant-go/swagger"
)

type App struct {
	cfg        *config.Config
	router     *httprouter.Router
	httpServer *http.Server
	pgxPool    *pgxpool.Pool
	minio      *minio.Client
}

func NewApp(ctx context.Context, cfg *config.Config) (App, error) {
	logging.GetLogger(ctx).Println("router init")
	router := httprouter.New()

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
		logging.GetLogger(ctx).Println("created minio S3 client")
	}

	return App{
		cfg:     cfg,
		router:  router,
		pgxPool: pgClient,
		minio:   minioClient,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	grp, ctx2 := errgroup.WithContext(ctx)
	grp.Go(func() error {
		return a.startHTTP(ctx2)
	})

	logging.GetLogger(ctx).Info("Application initialized and started")
	return grp.Wait()
}

func (a *App) startHTTP(ctx context.Context) error {
	controllerInit := controller.New(a.cfg, a.pgxPool, a.minio, a.router)
	errRoute := controllerInit.SetRoutes(ctx)
	if errRoute != nil {
		logging.GetLogger(ctx).WithError(errRoute).Fatal("failed to init routes")
	}

	logging.GetLogger(ctx).Printf("IP: %s, Port: %d", a.cfg.HTTP.Host, a.cfg.HTTP.Port)

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", a.cfg.HTTP.Host, a.cfg.HTTP.Port))
	if err != nil {
		logging.GetLogger(ctx).WithError(err).Fatal("failed to create http listener")
	}

	logging.GetLogger(ctx).Printf("CORS: %+v", a.cfg.Cors)

	c := cors.New(cors.Options{
		AllowedMethods:     strings.Split(a.cfg.Cors.AllowedMethods, ","),
		AllowedOrigins:     strings.Split(a.cfg.Cors.AllowedOrigins, ","),
		AllowedHeaders:     strings.Split(a.cfg.Cors.AllowedHeaders, ","),
		AllowCredentials:   a.cfg.Cors.AllowCredentials,
		OptionsPassthrough: a.cfg.Cors.OptionsPassthrough,
		ExposedHeaders:     strings.Split(a.cfg.Cors.ExposedHeaders, ","),
		Debug:              a.cfg.Cors.Debug,
	})

	hdl := c.Handler(a.router)

	a.httpServer = &http.Server{
		Handler:      hdl,
		WriteTimeout: a.cfg.HTTP.WriteTimeout,
		ReadTimeout:  a.cfg.HTTP.ReadTimeout,
	}

	logging.GetLogger(ctx).Println("http server started")
	if err = a.httpServer.Serve(listener); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			logging.GetLogger(ctx).Warningln("Server shutdown")
		default:
			logging.GetLogger(ctx).Fatalln(err)
		}
	}
	err = a.httpServer.Shutdown(context.Background())
	if err != nil {
		logging.GetLogger(ctx).Fatalln(err)
	}
	return err
}
