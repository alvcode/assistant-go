package handler

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/ucase"
	"assistant-go/internal/layer/vmodel"
	"assistant-go/internal/locale"
	"encoding/json"
	"fmt"
	"net/http"
)

type UserHandler struct {
	useCase ucase.UserUseCase
}

func NewUserHandler(useCase ucase.UserUseCase) *UserHandler {
	return &UserHandler{
		useCase: useCase,
	}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var createUserDto dto.UserLoginAndPassword
	langRequest := locale.GetLangFromContext(r.Context())

	err := json.NewDecoder(r.Body).Decode(&createUserDto)
	if err != nil {
		BlockEventHandle(r, BlockEventDecodeBodyType)
		SendErrorResponse(w, locale.T(langRequest, "error_reading_request_body"), http.StatusBadRequest, 0)
		return
	}
	if err := createUserDto.Validate(langRequest); err != nil {
		BlockEventHandle(r, BlockEventInputDataType)
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	entity, err := h.useCase.Create(r.Context(), createUserDto)
	if err != nil {
		BlockEventHandle(r, BlockEventOtherType)
		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}
	userVM := vmodel.UserFromEnity(entity)
	SendResponse(w, http.StatusCreated, userVM)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginUserDto dto.UserLoginAndPassword
	langRequest := locale.GetLangFromContext(r.Context())

	err := json.NewDecoder(r.Body).Decode(&loginUserDto)
	if err != nil {
		BlockEventHandle(r, BlockEventDecodeBodyType)
		SendErrorResponse(w, locale.T(langRequest, "error_reading_request_body"), http.StatusBadRequest, 0)
		return
	}
	if err := loginUserDto.Validate(langRequest); err != nil {
		BlockEventHandle(r, BlockEventInputDataType)
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}
	entity, err := h.useCase.Login(r.Context(), loginUserDto)
	if err != nil {
		BlockEventHandle(r, BlockEventErrorSignInType)
		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}

	userTokenVM := vmodel.UserTokenFromEnity(entity)
	SendResponse(w, http.StatusOK, userTokenVM)
}

func (h *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var refreshTokenDto dto.UserRefreshToken
	langRequest := locale.GetLangFromContext(r.Context())

	err := json.NewDecoder(r.Body).Decode(&refreshTokenDto)
	if err != nil {
		BlockEventHandle(r, BlockEventDecodeBodyType)
		SendErrorResponse(w, locale.T(langRequest, "error_reading_request_body"), http.StatusBadRequest, 0)
		return
	}
	if err := refreshTokenDto.Validate(langRequest); err != nil {
		BlockEventHandle(r, BlockEventInputDataType)
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	entity, err := h.useCase.RefreshToken(r.Context(), refreshTokenDto)
	if err != nil {
		BlockEventHandle(r, BlockEventRefreshTokenType)
		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}
	userTokenVM := vmodel.UserTokenFromEnity(entity)
	SendResponse(w, http.StatusOK, userTokenVM)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())

	authUser, err := GetAuthUser(r)
	if err != nil {
		BlockEventHandle(r, BlockEventUnauthorizedType)
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	err = h.useCase.Delete(r.Context(), authUser.ID)
	if err != nil {
		BlockEventHandle(r, BlockEventOtherType)
		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}

	SendResponse(w, http.StatusNoContent, nil)
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var userChangePasswordDto dto.UserChangePassword
	langRequest := locale.GetLangFromContext(r.Context())

	authUser, err := GetAuthUser(r)
	if err != nil {
		BlockEventHandle(r, BlockEventUnauthorizedType)
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&userChangePasswordDto)
	if err != nil {
		BlockEventHandle(r, BlockEventDecodeBodyType)
		SendErrorResponse(w, locale.T(langRequest, "error_reading_request_body"), http.StatusBadRequest, 0)
		return
	}

	if err := userChangePasswordDto.Validate(langRequest); err != nil {
		BlockEventHandle(r, BlockEventInputDataType)
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	err = h.useCase.ChangePassword(r.Context(), authUser.ID, userChangePasswordDto)
	if err != nil {
		BlockEventHandle(r, BlockEventOtherType)
		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}

	SendResponse(w, http.StatusNoContent, nil)
}
