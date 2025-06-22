package controller

import (
	"assistant-go/internal/config"
	"assistant-go/internal/handler"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/layer/ucase"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
	"github.com/minio/minio-go/v7"
	"net/http"
)

type Init struct {
	cfg    *config.Config
	db     *pgxpool.Pool
	minio  *minio.Client
	router *httprouter.Router
}

func New(cfg *config.Config, db *pgxpool.Pool, minio *minio.Client, router *httprouter.Router) *Init {
	return &Init{
		cfg:    cfg,
		db:     db,
		minio:  minio,
		router: router,
	}
}

func (controller *Init) SetRoutes(ctx context.Context) error {
	repos := repository.NewRepositories(ctx, controller.cfg, controller.db, controller.minio)

	handler.InitHandler(repos, controller.cfg)

	//controller.router.Handler(http.MethodGet, "/swagger", http.RedirectHandler("/swagger/index.html", http.StatusMovedPermanently))
	//controller.router.Handler(http.MethodGet, "/swagger/*any", httpSwagger.WrapHandler)

	controller.router.NotFound = handler.BuildHandler(handler.PageNotFoundHandler)
	controller.router.MethodNotAllowed = handler.BuildHandler(handler.PageNotFoundHandler)

	heartbeatHandler := handler.NewHeartbeatHandler()

	controller.router.Handler(
		http.MethodGet,
		"/api/heartbeat",
		handler.BuildHandler(heartbeatHandler.Heartbeat),
	)

	controller.setUserRoutes(ctx, repos)
	controller.setNotesCategories(ctx, repos)
	controller.setNotes(ctx, repos)
	controller.setFiles(ctx, repos)
	controller.setDrive(ctx, repos)

	return nil
}

func (controller *Init) setUserRoutes(ctx context.Context, repositories *repository.Repositories) {
	userUseCase := ucase.NewUserUseCase(ctx, repositories)
	userHandler := handler.NewUserHandler(userUseCase)

	controller.router.Handler(
		http.MethodPost,
		"/api/auth/register",
		handler.BuildHandler(userHandler.Create),
	)
	controller.router.Handler(
		http.MethodPost,
		"/api/auth/login",
		handler.BuildHandler(userHandler.Login),
	)
	controller.router.Handler(
		http.MethodPost,
		"/api/auth/refresh-token",
		handler.BuildHandler(userHandler.RefreshToken),
	)
	controller.router.Handler(
		http.MethodDelete,
		"/api/user",
		handler.BuildHandler(userHandler.Delete, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodPatch,
		"/api/user/change-password",
		handler.BuildHandler(userHandler.ChangePassword, handler.AuthMW),
	)
}

func (controller *Init) setNotesCategories(ctx context.Context, repositories *repository.Repositories) {
	noteCategoryUseCase := ucase.NewNoteCategoryUseCase(ctx, repositories)
	noteCategoryHandler := handler.NewNoteCategoryHandler(noteCategoryUseCase)

	controller.router.Handler(
		http.MethodPost,
		"/api/note-categories",
		handler.BuildHandler(noteCategoryHandler.Create, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodGet,
		"/api/note-categories",
		handler.BuildHandler(noteCategoryHandler.GetAll, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodDelete,
		"/api/note-categories/:id",
		handler.BuildHandler(noteCategoryHandler.Delete, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodPatch,
		"/api/note-categories/:id",
		handler.BuildHandler(noteCategoryHandler.Update, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodPost,
		"/api/note-categories/position-up",
		handler.BuildHandler(noteCategoryHandler.PositionUp, handler.AuthMW),
	)
}

func (controller *Init) setNotes(ctx context.Context, repositories *repository.Repositories) {
	noteUseCase := ucase.NewNoteUseCase(ctx, repositories)
	noteHandler := handler.NewNoteHandler(noteUseCase)

	controller.router.Handler(
		http.MethodPost,
		"/api/notes",
		handler.BuildHandler(noteHandler.Create, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodGet,
		"/api/notes",
		handler.BuildHandler(noteHandler.GetAll, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodPatch,
		"/api/notes",
		handler.BuildHandler(noteHandler.Update, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodGet,
		"/api/notes/:id",
		handler.BuildHandler(noteHandler.GetOne, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodDelete,
		"/api/notes/:id",
		handler.BuildHandler(noteHandler.DeleteOne, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodPost,
		"/api/notes/:id/pin",
		handler.BuildHandler(noteHandler.Pin, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodPost,
		"/api/notes/:id/unpin",
		handler.BuildHandler(noteHandler.UnPin, handler.AuthMW),
	)
}

func (controller *Init) setFiles(ctx context.Context, repositories *repository.Repositories) {
	fileUseCase := ucase.NewFileUseCase(ctx, repositories)
	fileHandler := handler.NewFileHandler(fileUseCase)

	controller.router.Handler(
		http.MethodPost,
		"/api/files",
		handler.BuildHandler(fileHandler.Upload, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodGet,
		"/api/files/hash/:hash",
		handler.BuildHandler(fileHandler.GetByHash),
	)
}

func (controller *Init) setDrive(ctx context.Context, repositories *repository.Repositories) {
	driveUseCase := ucase.NewDriveUseCase(ctx, repositories)
	driveHandler := handler.NewDriveHandler(driveUseCase)

	controller.router.Handler(
		http.MethodPost,
		"/api/drive/directories",
		handler.BuildHandler(driveHandler.CreateDirectory, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodGet,
		"/api/drive/tree",
		handler.BuildHandler(driveHandler.GetTree, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodGet,
		"/api/drive/files/:id",
		handler.BuildHandler(driveHandler.GetFile, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodPost,
		"/api/drive/upload-file",
		handler.BuildHandler(driveHandler.UploadFile, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodDelete,
		"/api/drive/:id",
		handler.BuildHandler(driveHandler.Delete, handler.AuthMW),
	)
	controller.router.Handler(
		http.MethodPatch,
		"/api/drive/files/:id/rename",
		handler.BuildHandler(driveHandler.Rename, handler.AuthMW),
	)
}
