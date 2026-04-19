package handler

import (
	"assistant-go/internal/layer/ucase"
	"assistant-go/internal/layer/vmodel"
	"assistant-go/internal/locale"
	"net/http"
)

type DriveRecycleBinHandler struct {
	useCase ucase.DriveRecycleBinUseCase
}

func NewDriveRecycleBinHandler(useCase ucase.DriveRecycleBinUseCase) *DriveRecycleBinHandler {
	return &DriveRecycleBinHandler{
		useCase: useCase,
	}
}

func (h *DriveRecycleBinHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())
	authUser, err := GetAuthUser(r)
	if err != nil {
		BlockEventHandle(r, BlockEventUnauthorizedType)
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	entities, err := h.useCase.GetAll(r.Context(), authUser)
	if err != nil {
		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}

	result := make([]*vmodel.DriveRecycleBinStruct, 0)
	for _, entity := range entities {
		result = append(result, vmodel.DriveRecycleBinStructFromEntity(entity))
	}
	SendResponse(w, http.StatusOK, result)
}

func (h *DriveRecycleBinHandler) RestoreOne(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())
	authUser, err := GetAuthUser(r)
	if err != nil {
		BlockEventHandle(r, BlockEventUnauthorizedType)
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	recycleBinID, err := getPathParamInt(r, "id")
	if err != nil {
		BlockEventHandle(r, BlockEventInputDataType)
		SendErrorResponse(w, locale.T(langRequest, "parameter_conversion_error"), http.StatusBadRequest, 0)
		return
	}

	err = h.useCase.RestoreOne(r.Context(), authUser, recycleBinID)
	if err != nil {
		SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		return
	}

	SendResponse(w, http.StatusNoContent, nil)
}
