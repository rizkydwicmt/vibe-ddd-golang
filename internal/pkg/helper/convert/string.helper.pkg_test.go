package convert

import (
	"testing"
)

func TestStringToStruct(t *testing.T) {
	type Dummy struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	jsonStr := `{"name":"John","age":30}`
	result, err := StringToStruct[Dummy](jsonStr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result == nil || result.Name != "John" || result.Age != 30 {
		t.Errorf("Unexpected result: %+v", result)
	}
}

func TestStringToJSON(t *testing.T) {
	jsonStr := `{"foo":"bar","num":123}`
	result, err := StringToJSON(jsonStr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result["foo"] != "bar" || int(result["num"].(float64)) != 123 {
		t.Errorf("Unexpected result: %+v", result)
	}
}

func TestStringToInt(t *testing.T) {
	res, err := StringToInt("42")
	if err != nil || res != 42 {
		t.Errorf("Expected 42, got %d, err: %v", res, err)
	}
	_, err = StringToInt("abc")
	if err == nil {
		t.Error("Expected error for non-integer string")
	}
}

func TestStringToInt64(t *testing.T) {
	res, err := StringToInt64("1234567890")
	if err != nil || res != 1234567890 {
		t.Errorf("Expected 1234567890, got %d, err: %v", res, err)
	}
	_, err = StringToInt64("abc")
	if err == nil {
		t.Error("Expected error for non-integer string")
	}
}

func TestGetMapStringValue(t *testing.T) {
	head := map[string]interface{}{"foo": "bar", "num": 123}
	val := GetMapStringValue(head, "foo")
	if val == nil || *val != "bar" {
		t.Errorf("Expected 'bar', got %v", val)
	}
	val = GetMapStringValue(head, "num")
	if val == nil || *val != "123" {
		t.Errorf("Expected '123', got %v", val)
	}
	val = GetMapStringValue(head, "notfound")
	if val == nil || *val != "" {
		t.Errorf("Expected empty string, got %v", val)
	}
}

func TestGetMapStringNilValue(t *testing.T) {
	head := map[string]interface{}{"foo": "bar", "num": 123}
	val := GetMapStringNilValue(head, "foo")
	if val == nil || *val != "bar" {
		t.Errorf("Expected 'bar', got %v", val)
	}
	val = GetMapStringNilValue(head, "num")
	if val == nil || *val != "123" {
		t.Errorf("Expected '123', got %v", val)
	}
	val = GetMapStringNilValue(head, "notfound")
	if val != nil {
		t.Errorf("Expected nil, got %v", val)
	}
}

func TestStringToBool(t *testing.T) {
	if !StringToBool("true") {
		t.Error("Expected true for 'true'")
	}
	if StringToBool("false") {
		t.Error("Expected false for 'false'")
	}
	if StringToBool("notabool") {
		t.Error("Expected false for invalid bool string")
	}
}
