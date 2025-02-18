package handler

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/ucase"
	"assistant-go/internal/locale"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

type NoteCategoryHandler struct {
	useCase ucase.NoteCategoryUseCase
}

func NewNoteCategoryHandler(useCase ucase.NoteCategoryUseCase) *NoteCategoryHandler {
	return &NoteCategoryHandler{
		useCase: useCase,
	}
}

func (h *NoteCategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())
	var createNoteCategoryDto dto.NoteCategoryCreate

	authUser, err := GetAuthUser(r)
	if err != nil {
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&createNoteCategoryDto)
	if err != nil {
		SendErrorResponse(w, locale.T(langRequest, "error_reading_request_body"), http.StatusBadRequest, 0)
		return
	}

	if err := createNoteCategoryDto.Validate(langRequest); err != nil {
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	entity, err := h.useCase.Create(createNoteCategoryDto, authUser, langRequest)
	if err != nil {
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}
	SendResponse(w, http.StatusCreated, entity)
}

func (h *NoteCategoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())

	authUser, err := GetAuthUser(r)
	if err != nil {
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	entities, err := h.useCase.FindAll(authUser.ID, langRequest)
	if err != nil {
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	SendResponse(w, http.StatusOK, entities)
}

func (h *NoteCategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())
	var idDto dto.RequiredID

	params := httprouter.ParamsFromContext(r.Context())

	if catIDStr := params.ByName("id"); catIDStr != "" {
		catIdInt, err := strconv.Atoi(catIDStr)

		if err != nil {
			SendErrorResponse(w, locale.T(langRequest, "parameter_conversion_error"), http.StatusBadRequest, 0)
		}
		idDto.ID = catIdInt
	}

	if err := idDto.Validate(langRequest); err != nil {
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	authUser, err := GetAuthUser(r)
	if err != nil {
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	err = h.useCase.Delete(authUser.ID, idDto.ID, langRequest)
	if err != nil {
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}
	SendResponse(w, http.StatusNoContent, nil)
}
