package validation

import "github.com/go-playground/validator/v10"

func validateMapStringString(fl validator.FieldLevel) bool {
	m, ok := fl.Field().Interface().(map[string]string)
	if !ok {
		return false
	}

	if len(m) == 0 {
		return false
	}

	for k, v := range m {
		if k == "" {
			return false
		}

		if v == "" {
			return false
		}
	}

	return true
}
