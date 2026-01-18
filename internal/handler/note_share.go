package handler

import (
	"assistant-go/internal/layer/ucase"
	"assistant-go/internal/locale"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

type NoteShareHandler struct {
	useCase ucase.NoteShareUseCase
}

func NewNoteShareHandler(useCase ucase.NoteShareUseCase) *NoteShareHandler {
	return &NoteShareHandler{
		useCase: useCase,
	}
}

func (h *NoteShareHandler) Create(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())

	authUser, err := GetAuthUser(r)
	if err != nil {
		BlockEventHandle(r, BlockEventUnauthorizedType)
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	var noteID int
	params := httprouter.ParamsFromContext(r.Context())
	if noteIDStr := params.ByName("id"); noteIDStr != "" {
		noteIDInt, err := strconv.Atoi(noteIDStr)

		if err != nil {
			BlockEventHandle(r, BlockEventInputDataType)
			SendErrorResponse(w, locale.T(langRequest, "parameter_conversion_error"), http.StatusBadRequest, 0)
			return
		}
		noteID = noteIDInt
	}

	noteShare, err := h.useCase.Create(r.Context(), noteID, authUser)
	if err != nil {
		BlockEventHandle(r, BlockEventOtherType)
		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}

	SendResponse(w, http.StatusCreated, noteShare)
}
