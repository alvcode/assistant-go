package handler

import (
	"assistant-go/internal/layer/dto/noteCategory"
	"assistant-go/internal/layer/useCase"
	"assistant-go/internal/locale"
	"encoding/json"
	"fmt"
	"net/http"
)

type NoteCategoryHandler struct {
	useCase useCase.NoteCategoryUseCase
}

func NewNoteCategoryHandler(useCase useCase.NoteCategoryUseCase) *NoteCategoryHandler {
	return &NoteCategoryHandler{
		useCase: useCase,
	}
}

func (h *NoteCategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())
	var createNoteCategoryDto dtoNoteCategory.Create

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
