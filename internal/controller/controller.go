package controller

import (
	"assistant-go/internal/config"
	"assistant-go/internal/handler"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/layer/useCase"
	"assistant-go/internal/logging"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
)

type Init struct {
	cfg    *config.Config
	db     *pgxpool.Pool
	router *httprouter.Router
}

func New(cfg *config.Config, db *pgxpool.Pool, router *httprouter.Router) *Init {
	return &Init{
		cfg:    cfg,
		db:     db,
		router: router,
	}
}

func (controller *Init) SetRoutes(ctx context.Context) error {
	logging.GetLogger(ctx).Println("swagger init")
	controller.router.Handler(http.MethodGet, "/swagger", http.RedirectHandler("/swagger/index.html", http.StatusMovedPermanently))
	controller.router.Handler(http.MethodGet, "/swagger/*any", httpSwagger.WrapHandler)

	heartbeatHandler := handler.NewHeartbeatHandler()
	controller.router.HandlerFunc(http.MethodGet, "/api/heartbeat", heartbeatHandler.Heartbeat)

	controller.setUserRoutes(ctx)

	return nil
}

func (controller *Init) setUserRoutes(ctx context.Context) {
	userRepository := repository.NewUserRepository(ctx, controller.db)
	userUseCase := useCase.NewUserUseCase(ctx, userRepository)
	userHandler := handler.NewUserHandler(userUseCase)

	controller.router.Handler(
		http.MethodPost,
		"/api/user/register",
		handler.LocaleMiddleware(userHandler.Create),
	)
}
