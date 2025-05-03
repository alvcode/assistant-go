package locale

import (
	"assistant-go/internal/logging"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var (
	bundleInstance *i18n.Bundle
	localizers     map[string]*i18n.Localizer
	once           sync.Once
)

func InitLocales(ctx context.Context) {
	once.Do(func() {
		bundleInstance = i18n.NewBundle(language.English)
		bundleInstance.RegisterUnmarshalFunc("json", json.Unmarshal)

		loadTranslations()

		localizers = make(map[string]*i18n.Localizer)
		localizers["en"] = i18n.NewLocalizer(bundleInstance, "en")
		localizers["ru"] = i18n.NewLocalizer(bundleInstance, "ru")

		logging.GetLogger(ctx).Infoln("Localization initialized")
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
			log.Fatalf("Failed to load translation file %s: %v", f.file, err)
		}
	}
}

func T(lang string, messageID string, args ...interface{}) string {
	fmt.Println(messageID, args)
	if localizers == nil {
		log.Printf("Localizers not initialized, returning messageID: %s", messageID)
		return messageID
	}
	localizer, exists := localizers[lang]
	if !exists {
		localizer = localizers["en"]
	}

	config := &i18n.LocalizeConfig{
		MessageID: messageID,
	}

	if len(args) > 0 {
		if templateData, ok := args[0].(map[string]interface{}); ok {
			config.TemplateData = templateData
		}
	}

	result, err := localizer.Localize(config)
	if err != nil {
		log.Printf("Translation not found: %s (%s)", messageID, lang)
		return messageID
	}

	return result
}

const localeContextKey = "locale"

var allowedLocales = map[string]bool{
	"ru": true,
	"en": true,
}

func Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		locale := r.Header.Get(localeContextKey)

		if !allowedLocales[locale] {
			locale = "en"
		}
		ctx := context.WithValue(r.Context(), localeContextKey, locale)
		next(w, r.WithContext(ctx))
	}
}

func GetLangFromContext(ctx context.Context) string {
	locale, ok := ctx.Value(localeContextKey).(string)
	if !ok {
		return "en"
	}
	return locale
}
