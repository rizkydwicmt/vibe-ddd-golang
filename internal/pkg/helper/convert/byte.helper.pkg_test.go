package convert

import (
	"reflect"
	"testing"
)

func TestByteToJSON(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		input := []byte(`{"foo": "bar", "num": 123}`)
		result, err := ByteToJSON(input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result["foo"] != "bar" || int(result["num"].(float64)) != 123 {
			t.Errorf("unexpected result: %v", result)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		input := []byte(`{"foo": "bar"`)
		_, err := ByteToJSON(input)
		if err == nil {
			t.Error("expected error for invalid JSON, got nil")
		}
	})

	t.Run("empty input", func(t *testing.T) {
		input := []byte("")
		_, err := ByteToJSON(input)
		if err == nil {
			t.Error("expected error for empty input, got nil")
		}
	})
}

func TestByteToStruct(t *testing.T) {
	type Sample struct {
		Foo string `json:"foo"`
		Num int    `json:"num"`
	}

	t.Run("valid struct", func(t *testing.T) {
		input := []byte(`{"foo": "bar", "num": 123}`)
		result, err := ByteToStruct[Sample](input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expected := &Sample{Foo: "bar", Num: 123}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}
	})

	t.Run("invalid struct", func(t *testing.T) {
		input := []byte(`{"foo": "bar"`)
		_, err := ByteToStruct[Sample](input)
		if err == nil {
			t.Error("expected error for invalid JSON, got nil")
		}
	})

	t.Run("empty input", func(t *testing.T) {
		input := []byte("")
		_, err := ByteToStruct[Sample](input)
		if err == nil {
			t.Error("expected error for empty input, got nil")
		}
	})

	t.Run("extra fields in JSON", func(t *testing.T) {
		input := []byte(`{"foo": "bar", "num": 123, "extra": "field"}`)
		result, err := ByteToStruct[Sample](input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expected := &Sample{Foo: "bar", Num: 123}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}
	})
}
