package handler

import (
	dtoNoteCategory "assistant-go/internal/layer/dto/noteCategory"
	"assistant-go/internal/layer/useCase"
	"assistant-go/internal/locale"
	"encoding/json"
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

	err := json.NewDecoder(r.Body).Decode(&createNoteCategoryDto)
	if err != nil {
		SendErrorResponse(w, locale.T(langRequest, "error_reading_request_body"), http.StatusBadRequest, 0)
		return
	}
	// TODO: продолжить...
}
