package convert

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func StringToStruct[I any](payload string) (result *I, err error) {
	err = json.Unmarshal([]byte(payload), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func TypeToString[T any](value T) string {
	return fmt.Sprintf("%v", value)
}

func StringToJSON(payload string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(payload), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func StringToInt(payload string) (int, error) {
	result, err := strconv.Atoi(payload)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func StringToInt64(payload string) (int64, error) {
	result, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func GetMapStringValue(header map[string]interface{}, key string) *string {
	str := ""
	value, exists := header[key]
	if !exists || value == nil {
		return &str
	}
	str = fmt.Sprintf("%v", value)
	return &str
}

func GetMapStringNilValue(header map[string]interface{}, key string) *string {
	str := ""
	value, exists := header[key]
	if !exists || value == nil {
		return nil
	}
	str = fmt.Sprintf("%v", value)
	return &str
}

func StringToBool(payload string) bool {
	value, err := strconv.ParseBool(payload)
	if err != nil {
		return false
	}
	return value
}
