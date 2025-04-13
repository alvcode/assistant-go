package handler

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/ucase"
	"assistant-go/internal/layer/vmodel"
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
		BlockEventHandle(r, BlockEventUnauthorizedType)
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&createNoteCategoryDto)
	if err != nil {
		BlockEventHandle(r, BlockEventDecodeBodyType)
		SendErrorResponse(w, locale.T(langRequest, "error_reading_request_body"), http.StatusBadRequest, 0)
		return
	}

	if err := createNoteCategoryDto.Validate(langRequest); err != nil {
		BlockEventHandle(r, BlockEventInputDataType)
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	entity, err := h.useCase.Create(createNoteCategoryDto, authUser)
	if err != nil {
		BlockEventHandle(r, BlockEventOtherType)
		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}

	result := vmodel.NoteCategoryFromEnity(entity)
	SendResponse(w, http.StatusCreated, result)
}

func (h *NoteCategoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())

	authUser, err := GetAuthUser(r)
	if err != nil {
		BlockEventHandle(r, BlockEventUnauthorizedType)
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	entities, err := h.useCase.FindAll(authUser.ID)
	if err != nil {
		BlockEventHandle(r, BlockEventOtherType)
		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}

	result := vmodel.NoteCategoriesFromEntities(entities)
	SendResponse(w, http.StatusOK, result)
}

func (h *NoteCategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())
	var idDto dto.RequiredID

	params := httprouter.ParamsFromContext(r.Context())

	if catIDStr := params.ByName("id"); catIDStr != "" {
		catIdInt, err := strconv.Atoi(catIDStr)

		if err != nil {
			BlockEventHandle(r, BlockEventInputDataType)
			SendErrorResponse(w, locale.T(langRequest, "parameter_conversion_error"), http.StatusBadRequest, 0)
			return
		}
		idDto.ID = catIdInt
	}

	if err := idDto.Validate(langRequest); err != nil {
		BlockEventHandle(r, BlockEventInputDataType)
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	authUser, err := GetAuthUser(r)
	if err != nil {
		BlockEventHandle(r, BlockEventUnauthorizedType)
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	err = h.useCase.Delete(authUser.ID, idDto.ID)
	if err != nil {
		BlockEventHandle(r, BlockEventOtherType)
		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}
	SendResponse(w, http.StatusNoContent, nil)
}

func (h *NoteCategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())
	var updateNoteCategoryDto dto.NoteCategoryUpdate

	params := httprouter.ParamsFromContext(r.Context())

	err := json.NewDecoder(r.Body).Decode(&updateNoteCategoryDto)
	if err != nil {
		BlockEventHandle(r, BlockEventDecodeBodyType)
		SendErrorResponse(w, locale.T(langRequest, "error_reading_request_body"), http.StatusBadRequest, 0)
		return
	}

	if catIDStr := params.ByName("id"); catIDStr != "" {
		catIdInt, err := strconv.Atoi(catIDStr)

		if err != nil {
			BlockEventHandle(r, BlockEventInputDataType)
			SendErrorResponse(w, locale.T(langRequest, "parameter_conversion_error"), http.StatusBadRequest, 0)
		}
		updateNoteCategoryDto.ID = catIdInt
	}

	if err := updateNoteCategoryDto.Validate(langRequest); err != nil {
		BlockEventHandle(r, BlockEventInputDataType)
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	authUser, err := GetAuthUser(r)
	if err != nil {
		BlockEventHandle(r, BlockEventUnauthorizedType)
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	entity, err := h.useCase.Update(updateNoteCategoryDto, authUser.ID)
	if err != nil {
		BlockEventHandle(r, BlockEventOtherType)
		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}

	result := vmodel.NoteCategoryFromEnity(entity)
	SendResponse(w, http.StatusOK, result)
}

func (h *NoteCategoryHandler) PositionUp(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())
	var categoryIDDto dto.RequiredID

	authUser, err := GetAuthUser(r)
	if err != nil {
		BlockEventHandle(r, BlockEventUnauthorizedType)
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&categoryIDDto)
	if err != nil {
		BlockEventHandle(r, BlockEventDecodeBodyType)
		SendErrorResponse(w, locale.T(langRequest, "error_reading_request_body"), http.StatusBadRequest, 0)
		return
	}

	if err := categoryIDDto.Validate(langRequest); err != nil {
		BlockEventHandle(r, BlockEventInputDataType)
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	err = h.useCase.PositionUp(categoryIDDto, authUser.ID)
	if err != nil {
		BlockEventHandle(r, BlockEventOtherType)
		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}

	SendResponse(w, http.StatusNoContent, nil)
}
