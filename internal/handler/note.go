package handler

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/ucase"
	"assistant-go/internal/layer/vmodel"
	"assistant-go/internal/locale"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type NoteHandler struct {
	useCase ucase.NoteUseCase
}

func NewNoteHandler(useCase ucase.NoteUseCase) *NoteHandler {
	return &NoteHandler{
		useCase: useCase,
	}
}

func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())
	var createNoteDto dto.NoteCreate

	authUser, err := GetAuthUser(r)
	if err != nil {
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&createNoteDto)
	if err != nil {
		SendErrorResponse(w, locale.T(langRequest, "error_reading_request_body"), http.StatusBadRequest, 0)
		return
	}

	if err = createNoteDto.Validate(langRequest); err != nil {
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	note, err := h.useCase.Create(createNoteDto, authUser, langRequest)
	if err != nil {
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	noteVModel := vmodel.NoteFromEnity(note)
	SendResponse(w, http.StatusCreated, noteVModel)
}

func (h *NoteHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())

	authUser, err := GetAuthUser(r)
	if err != nil {
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	var categoryID dto.RequiredID

	catIDStr := r.URL.Query().Get("categoryId")

	if catIDStr != "" {
		catIDInt, err := strconv.Atoi(catIDStr)

		if err != nil {
			SendErrorResponse(w, locale.T(langRequest, "parameter_conversion_error"), http.StatusBadRequest, 0)
			return
		}
		categoryID.ID = catIDInt
	}

	if err := categoryID.Validate(langRequest); err != nil {
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	notes, err := h.useCase.GetAll(categoryID, authUser, langRequest)
	if err != nil {
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}
	result := vmodel.NotesMinimalFromEntities(notes)
	SendResponse(w, http.StatusOK, result)
}

func (h *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())
	var updateNoteDto dto.NoteUpdate

	authUser, err := GetAuthUser(r)
	if err != nil {
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&updateNoteDto)
	if err != nil {
		SendErrorResponse(w, locale.T(langRequest, "error_reading_request_body"), http.StatusBadRequest, 0)
		return
	}

	if err = updateNoteDto.Validate(langRequest); err != nil {
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	note, err := h.useCase.Update(updateNoteDto, authUser, langRequest)
	if err != nil {
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	noteVModel := vmodel.NoteFromEnity(note)
	SendResponse(w, http.StatusCreated, noteVModel)
}

func (h *NoteHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())

	authUser, err := GetAuthUser(r)
	if err != nil {
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	var noteID dto.RequiredID

	noteIDStr := r.URL.Query().Get("id")

	if noteIDStr != "" {
		noteIDInt, err := strconv.Atoi(noteIDStr)

		if err != nil {
			SendErrorResponse(w, locale.T(langRequest, "parameter_conversion_error"), http.StatusBadRequest, 0)
			return
		}
		noteID.ID = noteIDInt
	}

	if err := noteID.Validate(langRequest); err != nil {
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}

	note, err := h.useCase.GetOne(noteID, authUser, langRequest)
	if err != nil {
		SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
		return
	}
	result := vmodel.NoteFromEnity(note)
	SendResponse(w, http.StatusOK, result)
}
