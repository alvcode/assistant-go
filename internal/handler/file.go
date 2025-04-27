package handler

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/ucase"
	"assistant-go/internal/locale"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
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
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			http.Error(w, "failed to close uploaded file", http.StatusInternalServerError)
			return
		}
	}(file)

	uploadFileDto := dto.UploadFile{
		File:             file,
		OriginalFilename: header.Filename,
		MaxSizeBytes:     appConf.File.UploadMaxSize << 20,
	}

	upload, err := h.useCase.Upload(uploadFileDto, authUser)
	if err != nil {
		return
	}

	maxUploadSize := appConf.File.UploadMaxSize << 20

	var allowedMimeTypes = map[string]string{
		"image/jpeg":      ".jpeg",
		"image/png":       ".png",
		"image/gif":       ".gif",
		"application/pdf": ".pdf",
		"application/zip": ".zip",
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	mimeType := http.DetectContentType(buffer)
	ext, allowed := allowedMimeTypes[mimeType]
	if !allowed {
		http.Error(w, "Invalid file type", http.StatusBadRequest)
		return
	}

	if seeker, ok := file.(io.Seeker); ok {
		_, err := seeker.Seek(0, io.SeekStart)
		if err != nil {
			http.Error(w, "Error resetting file pointer", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Unable to seek file", http.StatusInternalServerError)
		return
	}

	fileExt := strings.ToLower(filepath.Ext(header.Filename))
	if fileExt != ext {
		http.Error(w, "File extension doesn't match content type", http.StatusBadRequest)
		return
	}

	safeName := filepath.Base(header.Filename)
	if strings.Contains(safeName, "..") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	uploadPath := "./uploads"
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		err := os.Mkdir(uploadPath, 0755)
		if err != nil {
			http.Error(w, "Error create directory", http.StatusInternalServerError)
			return
		}
	}

	newFilename := fmt.Sprintf("file-%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadPath, newFilename)

	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			http.Error(w, "failed to close output file", http.StatusInternalServerError)
			return
		}
	}(out)

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

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
