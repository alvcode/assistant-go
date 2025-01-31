package handler

import (
	dtoUser "assistant-go/internal/layer/dto/user"
	"assistant-go/internal/layer/useCase"
	"assistant-go/internal/layer/viewModel"
	"encoding/json"
	"fmt"
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
	var createUserDto dtoUser.CreateDto

	err := json.NewDecoder(r.Body).Decode(&createUserDto)
	if err != nil {
		SendErrorResponse(w, "Error reading request body", http.StatusBadRequest, 0)
		return
	}

	if err := createUserDto.Validate(); err != nil {
		SendErrorResponse(w, fmt.Sprintf("Validation error: %v", err), http.StatusUnprocessableEntity, 0)
		return
	}

	entity, err := h.useCase.Create(createUserDto)
	if err != nil {
		return
	}

	user := viewModel.User{}
	userVM := user.FromEntity(entity)

	SendResponse(w, http.StatusCreated, userVM)
}
