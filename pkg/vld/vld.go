package vld

import (
	"assistant-go/internal/logging"
	"context"
	"errors"
	"fmt"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/ru"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	ru_translations "github.com/go-playground/validator/v10/translations/ru"
	"sync"
)

var (
	once       sync.Once
	translator *ut.UniversalTranslator
	Validate   *validator.Validate
)

func InitValidator(ctx context.Context) {
	once.Do(func() {
		enLocale := en.New()
		ruLocale := ru.New()
		translator = ut.New(enLocale, ruLocale)

		transEn, _ := translator.GetTranslator("en")
		transRu, _ := translator.GetTranslator("ru")

		Validate = validator.New()
		err := en_translations.RegisterDefaultTranslations(Validate, transEn)
		if err != nil {
			logging.GetLogger(ctx).Fatalln(fmt.Errorf("en_translations error: %v", err))
		}
		err = ru_translations.RegisterDefaultTranslations(Validate, transRu)
		if err != nil {
			logging.GetLogger(ctx).Fatalln(fmt.Errorf("ru_translations error: %v", err))
		}

		logging.GetLogger(ctx).Infoln("Validator and locales initialized")
	})
}

func GetTranslator(lang string) (trans ut.Translator) {
	translator, _ := translator.GetTranslator(lang)
	return translator
}

func TextFromFirstError(err error, lang string) error {
	errs := err.(validator.ValidationErrors)
	return errors.New(errs[0].Translate(GetTranslator(lang)))
}
