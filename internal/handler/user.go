package handler

import (
	dtoUser "assistant-go/internal/layer/dto/user"
	"assistant-go/internal/layer/useCase"
	"assistant-go/internal/layer/viewModel/user"
	"assistant-go/internal/locale"
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

	localeFromContext := locale.GetLocaleFromContext(r.Context())

	err := json.NewDecoder(r.Body).Decode(&createUserDto)
	if err != nil {
		SendErrorResponse(w, locale.T(localeFromContext, "error_reading_request_body"), http.StatusBadRequest, 0)
		return
	}

	if err := createUserDto.Validate(localeFromContext); err != nil {
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	entity, err := h.useCase.Create(createUserDto, localeFromContext)
	if err != nil {
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	userVM := vmUser.UserVMFromEnity(entity)

	SendResponse(w, http.StatusCreated, userVM)
}
