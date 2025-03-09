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

	repos := repository.NewRepositories(ctx, controller.db)

	controller.setUserRoutes(ctx, repos)
	controller.setNotesCategories(ctx, repos)
	controller.setNotes(ctx, repos)

	return nil
}

func (controller *Init) setUserRoutes(ctx context.Context, repositories *repository.Repositories) {
	userUseCase := ucase.NewUserUseCase(ctx, repositories)
	userHandler := handler.NewUserHandler(userUseCase)

	controller.router.Handler(
		http.MethodPost,
		"/api/auth/register",
		handler.BuildHandler(userHandler.Create, handler.BlockIPMW, handler.LocaleMW),
	)
	controller.router.Handler(
		http.MethodPost,
		"/api/auth/login",
		handler.BuildHandler(userHandler.Login, handler.BlockIPMW, handler.LocaleMW),
	)
	controller.router.Handler(
		http.MethodPost,
		"/api/auth/refresh-token",
		handler.BuildHandler(userHandler.RefreshToken, handler.BlockIPMW, handler.LocaleMW),
	)
	controller.router.Handler(
		http.MethodDelete,
		"/api/user",
		handler.BuildHandler(userHandler.Delete, handler.BlockIPMW, handler.LocaleMW, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodPatch,
		"/api/user/change-password",
		handler.BuildHandler(userHandler.ChangePassword, handler.LocaleMW, handler.AuthMW),
	)
}

func (controller *Init) setNotesCategories(ctx context.Context, repositories *repository.Repositories) {
	noteCategoryUseCase := ucase.NewNoteCategoryUseCase(ctx, repositories)
	noteCategoryHandler := handler.NewNoteCategoryHandler(noteCategoryUseCase)

	controller.router.Handler(
		http.MethodPost,
		"/api/note-categories",
		handler.BuildHandler(noteCategoryHandler.Create, handler.BlockIPMW, handler.LocaleMW, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodGet,
		"/api/note-categories",
		handler.BuildHandler(noteCategoryHandler.GetAll, handler.BlockIPMW, handler.LocaleMW, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodDelete,
		"/api/note-categories/:id",
		handler.BuildHandler(noteCategoryHandler.Delete, handler.BlockIPMW, handler.LocaleMW, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodPatch,
		"/api/note-categories/:id",
		handler.BuildHandler(noteCategoryHandler.Update, handler.BlockIPMW, handler.LocaleMW, handler.AuthMW),
	)

}

func (controller *Init) setNotes(ctx context.Context, repositories *repository.Repositories) {
	noteUseCase := ucase.NewNoteUseCase(ctx, repositories)
	noteHandler := handler.NewNoteHandler(noteUseCase)

	controller.router.Handler(
		http.MethodPost,
		"/api/notes",
		handler.BuildHandler(noteHandler.Create, handler.BlockIPMW, handler.LocaleMW, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodGet,
		"/api/notes",
		handler.BuildHandler(noteHandler.GetAll, handler.BlockIPMW, handler.LocaleMW, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodPatch,
		"/api/notes",
		handler.BuildHandler(noteHandler.Update, handler.BlockIPMW, handler.LocaleMW, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodGet,
		"/api/notes/:id",
		handler.BuildHandler(noteHandler.GetOne, handler.BlockIPMW, handler.LocaleMW, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodDelete,
		"/api/notes/:id",
		handler.BuildHandler(noteHandler.DeleteOne, handler.BlockIPMW, handler.LocaleMW, handler.AuthMW),
	)
}
