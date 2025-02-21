package controller

import (
	"assistant-go/internal/config"
	"assistant-go/internal/handler"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/layer/ucase"
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

	handler.InitMiddleware(ctx, controller.db)

	controller.router.Handler(http.MethodGet, "/swagger", http.RedirectHandler("/swagger/index.html", http.StatusMovedPermanently))
	controller.router.Handler(http.MethodGet, "/swagger/*any", httpSwagger.WrapHandler)

	heartbeatHandler := handler.NewHeartbeatHandler()
	controller.router.HandlerFunc(http.MethodGet, "/api/heartbeat", heartbeatHandler.Heartbeat)

	controller.setUserRoutes(ctx)
	controller.setNotesCategories(ctx)
	controller.setNotes(ctx)

	return nil
}

func (controller *Init) setUserRoutes(ctx context.Context) {
	userRepository := repository.NewUserRepository(ctx, controller.db)
	userUseCase := ucase.NewUserUseCase(ctx, userRepository)
	userHandler := handler.NewUserHandler(userUseCase)

	controller.router.Handler(
		http.MethodPost,
		"/api/auth/register",
		handler.BuildHandler(userHandler.Create, handler.LocaleMW),
	)
	controller.router.Handler(
		http.MethodPost,
		"/api/auth/login",
		handler.BuildHandler(userHandler.Login, handler.LocaleMW),
	)
	controller.router.Handler(
		http.MethodPost,
		"/api/auth/refresh-token",
		handler.BuildHandler(userHandler.RefreshToken, handler.LocaleMW),
	)
}

func (controller *Init) setNotesCategories(ctx context.Context) {
	noteCategoryRepository := repository.NewNoteCategoryRepository(ctx, controller.db)
	noteCategoryUseCase := ucase.NewNoteCategoryUseCase(ctx, noteCategoryRepository)
	noteCategoryHandler := handler.NewNoteCategoryHandler(noteCategoryUseCase)

	controller.router.Handler(
		http.MethodPost,
		"/api/notes/categories",
		handler.BuildHandler(noteCategoryHandler.Create, handler.LocaleMW, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodGet,
		"/api/notes/categories",
		handler.BuildHandler(noteCategoryHandler.GetAll, handler.LocaleMW, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodDelete,
		"/api/notes/categories/:id",
		handler.BuildHandler(noteCategoryHandler.Delete, handler.LocaleMW, handler.AuthMW),
	)
}

func (controller *Init) setNotes(ctx context.Context) {
	noteRepository := repository.NewNoteRepository(ctx, controller.db)
	noteCategoryRepository := repository.NewNoteCategoryRepository(ctx, controller.db)
	noteUseCase := ucase.NewNoteUseCase(ctx, noteRepository, noteCategoryRepository)
	noteHandler := handler.NewNoteHandler(noteUseCase)

	controller.router.Handler(
		http.MethodPost,
		"/api/notes",
		handler.BuildHandler(noteHandler.Create, handler.LocaleMW, handler.AuthMW),
	)
}
