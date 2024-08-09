package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

func NewValidator() (v *validator.Validate) {
	v = validator.New()
	// register function to get tag name from json tags.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		v, ok := fld.Tag.Lookup("json")
		if !ok {
			v, _ = fld.Tag.Lookup("form")
			name := strings.SplitN(v, ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		}

		name := strings.SplitN(v, ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return
}

func NewTranslation(v *validator.Validate) ut.Translator {
	english := en.New()
	uni := ut.New(english, english)
	trans, _ := uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(v, trans)
	return trans
}
