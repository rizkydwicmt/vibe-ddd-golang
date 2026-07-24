package convert

import (
	"testing"
)

type Dummy struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestJSONToString(t *testing.T) {
	obj := Dummy{Name: "John", Age: 30}
	jsonStr, err := JSONToString(obj)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if jsonStr != `{"name":"John","age":30}` {
		t.Errorf("Unexpected JSON string: %s", jsonStr)
	}
}

func TestJSONToStruct(t *testing.T) {
	obj := Dummy{Name: "Jane", Age: 25}
	result, err := JSONToStruct[Dummy](obj)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result == nil || result.Name != "Jane" || result.Age != 25 {
		t.Errorf("Unexpected result: %+v", result)
	}
}

func TestJSONToByte(t *testing.T) {
	obj := Dummy{Name: "Alice", Age: 22}
	jsonBytes, err := JSONToByte(obj)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := []byte(`{"name":"Alice","age":22}`)
	if string(jsonBytes) != string(expected) {
		t.Errorf("Unexpected JSON bytes: %s", string(jsonBytes))
	}
}
