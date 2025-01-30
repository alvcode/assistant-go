package handler

import (
	"assistant-go/internal/layer/useCase"
	"net/http"
)

type UserHandler struct {
	useCase useCase.UserUseCase
}

func NewUserHandler(useCase useCase.UserUseCase) *UserHandler {
	return &UserHandler{
		useCase: useCase,
	}
}

// Create
// @Summary Heartbeat metric
// @Tags Metrics
// @Success 204
// @Failure 400
// @Router /api/user/register [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	// TODO: продолжить
	w.WriteHeader(http.StatusNoContent)
}
