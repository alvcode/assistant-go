package handler

import (
	dtoUser "assistant-go/internal/layer/dto/user"
	"assistant-go/internal/layer/useCase"
	"assistant-go/internal/layer/viewModel/user"
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
		SendErrorResponse(w, fmt.Sprintf("Create user error: %v", err), http.StatusUnprocessableEntity, 0)
		return
	}

	userVM := vmUser.UserVMFromEnity(entity)

	SendResponse(w, http.StatusCreated, userVM)
}
