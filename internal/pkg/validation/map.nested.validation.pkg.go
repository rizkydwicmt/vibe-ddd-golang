package validation

import "github.com/go-playground/validator/v10"

func validateNestedMap(fl validator.FieldLevel) bool {
	m, ok := fl.Field().Interface().(map[string]interface{})
	if !ok {
		return false
	}

	if len(m) == 0 {
		return false
	}

	return validateMapContent(m)
}

func validateMapContent(m map[string]interface{}) bool {
	for k, v := range m {
		if k == "" {
			return false
		}

		switch val := v.(type) {
		case string:
			if val == "" {
				return false
			}

		case float64:
			if val < 0 {
				return false
			}

		case bool:

		case []interface{}:
			for _, elem := range val {
				if !validateElement(elem) {
					return false
				}
			}

		case map[string]interface{}:
			if !validateMapContent(val) {
				return false
			}

		case nil:

		default:
			return false
		}
	}

	return true
}

func validateElement(elem interface{}) bool {
	switch val := elem.(type) {
	case string:
		return val != ""
	case float64:
		return val >= 0
	case bool:
		return true
	case []interface{}:
		for _, e := range val {
			if !validateElement(e) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		return validateMapContent(val)
	case nil:
		return false
	default:
		return false
	}
}
