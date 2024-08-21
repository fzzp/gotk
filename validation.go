package gotk

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

type RegisterValidationHandler func(validate *validator.Validate)

var (
	once     sync.Once
	trans    ut.Translator
	validate *validator.Validate
	phoneRX  = regexp.MustCompile(`^1[3-9]\d{9}$`)
)

var (
	ErrInvalidInput = fmt.Errorf("入参错误")
)

// InitValidation 初始化验证器，仅会执行一次，locale 取值范围[zh、en]之一，若不是默认zh
func InitValidation(locale string, customHandlers ...RegisterValidationHandler) (ut.Translator, *validator.Validate) {
	// 仅第一次调用的时候才会调用
	once.Do(func() {
		if !(locale == "zh" || locale == "en") {
			locale = "zh"
		}

		// 设置翻译语言支持
		en := en.New()
		zh := zh.New()
		uni := ut.New(en, en, zh)

		// 这里不需要做额外的处理，找不到zh就会使用会退的en
		var ok bool
		if trans, ok = uni.GetTranslator(locale); !ok {
			fmt.Printf("->>> uni.GetTranslator(%q) ok = %t \n", locale, ok)
		}

		validate = validator.New()

		// 自定义验证
		validate.RegisterValidation("vphone", func(fieldLevel validator.FieldLevel) bool {
			if val, ok := fieldLevel.Field().Interface().(string); ok {
				if phoneRX.MatchString(val) {
					return true
				}
			}
			return false
		})

		if len(customHandlers) > 0 {
			for _, handle := range customHandlers {
				handle(validate)
			}
		}

		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			// zh 自定义Tag struct { Email string `json:"email" zh:"邮箱地址"` }
			name := strings.SplitN(fld.Tag.Get("zh"), ",", 2)[0]
			if name == "-" {
				return ""
			}

			name = strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}

			name = strings.SplitN(fld.Tag.Get("form"), ",", 2)[0]
			if name == "-" {
				return ""
			}

			if name == "" {
				name = fld.Name
			}

			return name + " "
		})

		switch locale {
		case "zh":
			zh_translations.RegisterDefaultTranslations(validate, trans)
		default:
			en_translations.RegisterDefaultTranslations(validate, trans)
		}
	})

	return trans, validate
}

// CheckStruct 检查结构体是否通过验证，input 必须是结构体指针,
// 请先调用 InitValidation 初始化
func CheckStruct(input interface{}) error {
	if err := validate.Struct(input); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			// 这里一般是 input 不是指针
			fmt.Println("input 参数请使用结构体指针")
			return ErrInvalidInput
		}

		errs := err.(validator.ValidationErrors)

		var msgs []string

		// 错误消息翻译和拼接
		for _, e := range errs.Translate(trans) {
			msgs = append(msgs, e)
		}

		err = fmt.Errorf(strings.Join(msgs, ";"))
		return err
	}

	return nil
}
