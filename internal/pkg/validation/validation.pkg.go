package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"vibe-ddd-golang/internal/common/enum"
	types "vibe-ddd-golang/internal/common/type"

	"github.com/gin-gonic/gin/binding"

	"github.com/go-playground/validator/v10"
)

var val *validator.Validate

var validationMessages = map[string]string{
	"e164":         "must be a e164 formatted phone number",
	"required":     "is required",
	"url":          "must be a valid URL",
	"datetime":     "must be a valid date-time format (2006-01-02T15:04:05Z07:00)",
	"number":       "must be a number",
	"oneof":        "must be one of the allowed values: %s",
	"email":        "must be a valid email address",
	"min":          "must be greater than or equal to %s",
	"max":          "must be less than or equal to %s",
	"len":          "must have the exact length of %s",
	"alpha":        "must contain only alphabetic characters",
	"alphanum":     "must contain only alphanumeric characters",
	"eqfield":      "must be equal to the value of the %s field",
	"nefield":      "must not be equal to the value of the %s field",
	"gt":           "must be greater than %s",
	"gte":          "must be greater than or equal to %s",
	"lt":           "must be less than %s",
	"lte":          "must be less than or equal to %s",
	"excludes":     "must not contain the value %s",
	"excludesall":  "must not contain any of the values: %s",
	"enum":         "must be one of the allowed enum values: %s",
	"stringToBool": "must be a boolean value",
	"timezone":     "must be a valid timezone",
}

func Setup() error {
	val = validator.New(validator.WithRequiredStructEnabled())
	val.SetTagName("validate")

	if err := registerValidations(val); err != nil {
		return fmt.Errorf("failed to register custom validations: %w", err)
	}

	val.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.SetTagName("validate")
		if err := registerValidations(v); err != nil {
			return fmt.Errorf("failed to register custom validations in Gin engine: %w", err)
		}
	} else {
		return fmt.Errorf("failed to get validation engine")
	}

	return nil
}

func registerValidations(v *validator.Validate) error {
	if err := v.RegisterValidation("enum", enum.ValidateEnum); err != nil {
		return fmt.Errorf("failed to register enum validation: %w", err)
	}
	if err := v.RegisterValidation("stringToBool", types.ValidateStringToBool); err != nil {
		return fmt.Errorf("failed to register stringToBool validation: %w", err)
	}
	if err := v.RegisterValidation("mapStringString", validateMapStringString); err != nil {
		return fmt.Errorf("failed to register map string validation: %w", err)
	}
	if err := v.RegisterValidation("mapStringInterface", validateNestedMap); err != nil {
		return fmt.Errorf("failed to register nested map validation: %w", err)
	}
	if err := v.RegisterValidation("timezone", validateTimezone); err != nil {
		return fmt.Errorf("failed to register timezone validation: %w", err)
	}
	return nil
}

func validateTimezone(fl validator.FieldLevel) bool {
	timezone, ok := fl.Field().Interface().(string)
	if !ok || strings.TrimSpace(timezone) == "" {
		return false
	}
	_, err := time.LoadLocation(timezone)
	return err == nil
}

func Validate(payload interface{}) error {
	if val == nil {
		v, err := newValidator("validate")
		if err != nil {
			return err
		}
		val = v
	}
	return validateWith(val, payload)
}

func newValidator(tagName string) (*validator.Validate, error) {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.SetTagName(tagName)
	if err := registerValidations(v); err != nil {
		return nil, fmt.Errorf("failed to register custom validations: %w", err)
	}
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})
	return v, nil
}

func validateWith(v *validator.Validate, payload interface{}) error {
	if err := v.Struct(payload); err != nil {
		var errorMessages []string

		validationErrors := parsingErrorValidate(err)
		if validationErrors != "" {
			errorMessages = append(errorMessages, validationErrors)
		}
		message := "Validation failed: " + strings.Join(errorMessages, ", ")
		return errors.New(message)
	}

	return nil
}

func parsingErrorValidate(err error) string {
	var errs validator.ValidationErrors
	if errors.As(err, &errs) {
		var sb strings.Builder
		for _, e := range errs {
			name := e.Namespace()
			field := e.Field()
			tag := e.Tag()
			param := e.Param()
			tp := e.Type()

			msg := validationMessages[tag]
			switch tag {
			case "enum":
				msg = fmt.Sprintf(msg, tp)
			default:
				if strings.Contains(msg, "%s") {
					msg = fmt.Sprintf(msg, param)
				}
			}
			sb.WriteString(fmt.Sprintf("%s: %s %s", name, field, msg))
			sb.WriteString(", ")
		}
		return strings.TrimSuffix(sb.String(), ", ")
	}
	return err.Error()
}
