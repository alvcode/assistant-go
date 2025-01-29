package app

import (
	"assistant-go/internal/config"
	"assistant-go/internal/handler"
	"assistant-go/internal/logging"
	"assistant-go/internal/storage/postgres"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"time"

	_ "assistant-go/swagger"
)

type App struct {
	cfg        *config.Config
	router     *httprouter.Router
	httpServer *http.Server
	pgxPool    *pgxpool.Pool
}

func NewApp(ctx context.Context, cfg *config.Config) (App, error) {
	logging.GetLogger(ctx).Println("router init")
	router := httprouter.New()

	logging.GetLogger(ctx).Println("swagger init")
	router.Handler(http.MethodGet, "/swagger", http.RedirectHandler("/swagger/index.html", http.StatusMovedPermanently))
	router.Handler(http.MethodGet, "/swagger/*any", httpSwagger.WrapHandler)

	heartbeatHandler := handler.HeartbeatHandler{}
	heartbeatHandler.Register(router)

	pgConfig := postgres.NewPgConfig(cfg.DB.Host, cfg.DB.Port, cfg.DB.Username, cfg.DB.Password, cfg.DB.Database)
	pgClient, err := postgres.NewClient(ctx, 5, time.Second*5, pgConfig)
	if err != nil {
		logging.GetLogger(ctx).Fatalln(err)
	}

	return App{
		cfg:     cfg,
		router:  router,
		pgxPool: pgClient,
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
	logging.GetLogger(ctx).Printf("IP: %s, Port: %d", a.cfg.HTTP.Host, a.cfg.HTTP.Port)

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", a.cfg.HTTP.Host, a.cfg.HTTP.Port))
	if err != nil {
		logging.GetLogger(ctx).WithError(err).Fatal("failed to create http listener")
	}

	logging.GetLogger(ctx).Printf("CORS: %+v", a.cfg.Cors)

	c := cors.New(cors.Options{
		AllowedMethods:     a.cfg.Cors.AllowedMethods,
		AllowedOrigins:     a.cfg.Cors.AllowedOrigins,
		AllowedHeaders:     a.cfg.Cors.AllowedHeaders,
		AllowCredentials:   a.cfg.Cors.AllowCredentials,
		OptionsPassthrough: a.cfg.Cors.OptionsPassthrough,
		ExposedHeaders:     a.cfg.Cors.ExposedHeaders,
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
