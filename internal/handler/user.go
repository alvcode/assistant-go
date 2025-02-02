package handler

import (
	"assistant-go/internal/layer/dto/user"
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
	var createUserDto dtoUser.LoginAndPassword
	langRequest := locale.GetLangFromContext(r.Context())

	err := json.NewDecoder(r.Body).Decode(&createUserDto)
	if err != nil {
		SendErrorResponse(w, locale.T(langRequest, "error_reading_request_body"), http.StatusBadRequest, 0)
		return
	}
	if err := createUserDto.Validate(langRequest); err != nil {
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	entity, err := h.useCase.Create(createUserDto, langRequest)
	if err != nil {
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}
	userVM := vmUser.UserVMFromEnity(entity)
	SendResponse(w, http.StatusCreated, userVM)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginUserDto dtoUser.LoginAndPassword
	langRequest := locale.GetLangFromContext(r.Context())

	err := json.NewDecoder(r.Body).Decode(&loginUserDto)
	if err != nil {
		SendErrorResponse(w, locale.T(langRequest, "error_reading_request_body"), http.StatusBadRequest, 0)
		return
	}
	if err := loginUserDto.Validate(langRequest); err != nil {
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}
	entity, err := h.useCase.Login(loginUserDto, langRequest)
	if err != nil {
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	userTokenVM := vmUser.UserTokenVMFromEnity(entity)
	SendResponse(w, http.StatusOK, userTokenVM)
}
