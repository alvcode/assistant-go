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

type FileHandler struct {
	useCase ucase.FileUseCase
}

func NewFileHandler(useCase ucase.FileUseCase) *FileHandler {
	return &FileHandler{
		useCase: useCase,
	}
}

/*

// Допустимые MIME-типы и соответствующие расширения
var allowedMimeTypes = map[string]string{
	"image/jpeg": ".jpeg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"application/pdf": ".pdf",
}

// Максимальный размер файла (5MB)
const maxUploadSize = 5 << 20

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Проверка метода запроса
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Ограничение размера файла
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	// 3. Получение файла из формы
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 4. Проверка MIME-типа
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Возвращаем указатель чтения в начало файла
	_, err = file.Seek(0, 0)
	if err != nil {
		http.Error(w, "Error seeking file", http.StatusInternalServerError)
		return
	}

	mimeType := http.DetectContentType(buffer)
	ext, allowed := allowedMimeTypes[mimeType]
	if !allowed {
		http.Error(w, "Invalid file type", http.StatusBadRequest)
		return
	}

	// 5. Проверка расширения файла
	fileExt := strings.ToLower(filepath.Ext(header.Filename))
	if fileExt != ext {
		http.Error(w, "File extension doesn't match content type", http.StatusBadRequest)
		return
	}

	// 6. Создание папки для загрузок, если её нет
	uploadPath := "./uploads"
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		os.Mkdir(uploadPath, 0755)
	}

	// 7. Генерация безопасного имени файла
	// В реальном приложении лучше использовать UUID или хеш
	newFilename := fmt.Sprintf("file-%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadPath, newFilename)

	// 8. Сохранение файла
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	// 9. Ответ об успешной загрузке
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "File uploaded successfully: %s", newFilename)
}


*/

func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
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
