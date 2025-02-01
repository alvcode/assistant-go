package locale

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var (
	bundleInstance *i18n.Bundle
	localizers     map[string]*i18n.Localizer
	once           sync.Once
)

func InitLocales() {
	once.Do(func() {
		bundleInstance = i18n.NewBundle(language.English)
		bundleInstance.RegisterUnmarshalFunc("json", json.Unmarshal)

		loadTranslations()

		localizers = make(map[string]*i18n.Localizer)
		localizers["en"] = i18n.NewLocalizer(bundleInstance, "en")
		localizers["ru"] = i18n.NewLocalizer(bundleInstance, "ru")

		log.Println("Localization initialized")
	})
}

func loadTranslations() {
	files := []struct {
		lang string
		file string
	}{
		{"en", "internal/locale/messages/en/app.en.json"},
		{"en", "internal/locale/messages/en/error.en.json"},
		{"ru", "internal/locale/messages/ru/app.ru.json"},
		{"ru", "internal/locale/messages/ru/error.ru.json"},
	}

	for _, f := range files {
		if _, err := bundleInstance.LoadMessageFile(f.file); err != nil {
			log.Printf("Failed to load translation file %s: %v", f.file, err)
		}
	}
}

func T(lang string, messageID string, args ...interface{}) string {
	localizer, exists := localizers[lang]
	if !exists {
		localizer = localizers["en"]
	}

	result, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: args,
	})
	if err != nil {
		log.Printf("Translation not found: %s (%s)", messageID, lang)
		return messageID
	}

	return result
}

// Key для хранения локализатора в контексте запроса
type contextKey string

const localeContextKey contextKey = "locale"

func LocaleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем заголовок "locale" (например, "en" или "ru")
		localeHeader := r.Header.Get("locale")
		if localeHeader == "" {
			localeHeader = "en" // Если заголовок не задан — по умолчанию английский
		}

		// Преобразуем в нижний регистр для стабильности
		localeHeader = strings.ToLower(localeHeader)

		// Проверяем, поддерживаем ли данный язык
		if _, exists := localizers[localeHeader]; !exists {
			// Если язык не поддерживается — fallback на английский
			localeHeader = "en"
		}

		// Устанавливаем локализацию в контекст запроса
		ctx := context.WithValue(r.Context(), localeContextKey, localeHeader)

		// Применяем контекст с локализацией к следующему обработчику
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetLocaleFromContext(ctx context.Context) string {
	locale, ok := ctx.Value(localeContextKey).(string)
	if !ok {
		// Если локализация не установлена, возвращаем английский как fallback
		return "en"
	}
	return locale
}
