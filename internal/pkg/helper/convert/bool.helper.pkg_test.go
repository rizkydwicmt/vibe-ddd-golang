package convert

import (
	"testing"
)

func TestGetMapBoolValue(t *testing.T) {
	t.Run("valid true string", func(t *testing.T) {
		h := map[string]interface{}{"flag": "true"}
		result := GetMapBoolValue(h, "flag")
		if result == nil || *result != true {
			t.Errorf("expected true, got %v", result)
		}
	})

	t.Run("valid false string", func(t *testing.T) {
		h := map[string]interface{}{"flag": "false"}
		result := GetMapBoolValue(h, "flag")
		if result == nil || *result != false {
			t.Errorf("expected false, got %v", result)
		}
	})

	t.Run("case insensitive true", func(t *testing.T) {
		h := map[string]interface{}{"flag": "TrUe"}
		result := GetMapBoolValue(h, "flag")
		if result == nil || *result != true {
			t.Errorf("expected true, got %v", result)
		}
	})

	t.Run("invalid bool string", func(t *testing.T) {
		h := map[string]interface{}{"flag": "notabool"}
		result := GetMapBoolValue(h, "flag")
		if result == nil || *result != false {
			t.Errorf("expected false for invalid bool, got %v", result)
		}
	})

	t.Run("missing key", func(t *testing.T) {
		h := map[string]interface{}{"other": "true"}
		result := GetMapBoolValue(h, "flag")
		if result == nil || *result != false {
			t.Errorf("expected false for missing key, got %v", result)
		}
	})

	t.Run("non-string value", func(t *testing.T) {
		h := map[string]interface{}{"flag": 123}
		result := GetMapBoolValue(h, "flag")
		if result == nil || *result != false {
			t.Errorf("expected false for non-string value, got %v", result)
		}
	})
}
