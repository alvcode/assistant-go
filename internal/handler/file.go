package handler

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/ucase"
	"assistant-go/internal/locale"
	"mime/multipart"
	"net/http"
)

type FileHandler struct {
	useCase ucase.FileUseCase
}

func NewFileHandler(useCase ucase.FileUseCase) *FileHandler {
	return &FileHandler{
		useCase: useCase,
	}
}

/*

use case

type UploadFileInput struct {
	File multipart.File
	OriginalFilename string
	MaxSize int64
	AllowedMimeTypes map[string]string
}

type UploadFileOutput struct {
	NewFilename string
	Path        string
	Size        int64
}

func (u *FileUploader) Upload(input UploadFileInput) (UploadFileOutput, error) {
	// validate mime, extension, etc
	// create directory if needed
	// save file
	// return output
}


handler

func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// чтение файла через r.FormFile

	in := UploadFileInput{
		File: file,
		OriginalFilename: header.Filename,
		MaxSize: appConf.File.UploadMaxSize << 20,
		AllowedMimeTypes: map[string]string{...},
	}

	out, err := h.fileUploader.Upload(in)
	if err != nil {
		SendErrorResponse(w, err.Error(), http.StatusBadRequest, 0)
		return
	}

	// например, вернуть путь или имя файла
	json.NewEncoder(w).Encode(out)
}


*/

func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	langRequest := locale.GetLangFromContext(r.Context())
	//var uploadFileDto dto.UploadFile

	authUser, err := GetAuthUser(r)
	if err != nil {
		BlockEventHandle(r, BlockEventUnauthorizedType)
		SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		SendErrorResponse(w, "Invalid file", http.StatusUnauthorized, 0)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			SendErrorResponse(w, "failed to close uploaded file", http.StatusUnauthorized, 0)
			return
		}
	}(file)

	uploadFileDto := dto.UploadFile{
		File:             file,
		OriginalFilename: header.Filename,
		MaxSizeBytes:     appConf.File.UploadMaxSize << 20,
		SavePath:         appConf.File.SavePath,
	}

	upload, err := h.useCase.Upload(uploadFileDto, authUser)
	if err != nil {
		//SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
		SendErrorResponse(w, err.Error(), http.StatusUnprocessableEntity, 0)
		return
	}

	SendResponse(w, http.StatusCreated, upload)
	return

	//var createNoteCategoryDto dto.NoteCategoryCreate
	//
	//authUser, err := GetAuthUser(r)
	//if err != nil {
	//	BlockEventHandle(r, BlockEventUnauthorizedType)
	//	SendErrorResponse(w, locale.T(langRequest, "unauthorized"), http.StatusUnauthorized, 0)
	//	return
	//}
	//
	//err = json.NewDecoder(r.Body).Decode(&createNoteCategoryDto)
	//if err != nil {
	//	BlockEventHandle(r, BlockEventDecodeBodyType)
	//	SendErrorResponse(w, locale.T(langRequest, "error_reading_request_body"), http.StatusBadRequest, 0)
	//	return
	//}
	//
	//if err := createNoteCategoryDto.Validate(langRequest); err != nil {
	//	BlockEventHandle(r, BlockEventInputDataType)
	//	SendErrorResponse(w, fmt.Sprint(err), http.StatusUnprocessableEntity, 0)
	//	return
	//}
	//
	//entity, err := h.useCase.Create(createNoteCategoryDto, authUser)
	//if err != nil {
	//	BlockEventHandle(r, BlockEventOtherType)
	//	SendErrorResponse(w, buildErrorMessage(langRequest, err), http.StatusUnprocessableEntity, 0)
	//	return
	//}
	//
	//result := vmodel.NoteCategoryFromEnity(entity)
	//SendResponse(w, http.StatusCreated, result)
}
