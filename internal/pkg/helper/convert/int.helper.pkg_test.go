package convert

import (
	"reflect"
	"testing"
)

func TestGetMapIntValue(t *testing.T) {
	tests := []struct {
		name   string
		header map[string]interface{}
		key    string
		expect int
	}{
		{"int value", map[string]interface{}{"a": 5}, "a", 5},
		{"float value", map[string]interface{}{"a": 5.7}, "a", 5},
		{"uint value", map[string]interface{}{"a": uint(7)}, "a", 7},
		{"bool true", map[string]interface{}{"a": true}, "a", 1},
		{"bool false", map[string]interface{}{"a": false}, "a", 0},
		{"string int", map[string]interface{}{"a": "42"}, "a", 42},
		{"string non-int", map[string]interface{}{"a": "abc"}, "a", 0},
		{"missing key", map[string]interface{}{"b": 1}, "a", 0},
		{"nil value", map[string]interface{}{"a": nil}, "a", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetMapIntValue(tt.header, tt.key)
			if got == nil || *got != tt.expect {
				t.Errorf("%s: got %v, want %v", tt.name, got, tt.expect)
			}
		})
	}
}

func TestGetMapInt64Value(t *testing.T) {
	tests := []struct {
		name   string
		header map[string]interface{}
		key    string
		expect int64
	}{
		{"int value", map[string]interface{}{"a": int64(5)}, "a", 5},
		{"float value", map[string]interface{}{"a": 5.7}, "a", 5},
		{"uint value", map[string]interface{}{"a": uint64(7)}, "a", 7},
		{"bool true", map[string]interface{}{"a": true}, "a", 1},
		{"bool false", map[string]interface{}{"a": false}, "a", 0},
		{"string int", map[string]interface{}{"a": "42"}, "a", 42},
		{"string non-int", map[string]interface{}{"a": "abc"}, "a", 0},
		{"missing key", map[string]interface{}{"b": 1}, "a", 0},
		{"nil value", map[string]interface{}{"a": nil}, "a", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetMapInt64Value(tt.header, tt.key)
			if got == nil || *got != tt.expect {
				t.Errorf("%s: got %v, want %v", tt.name, got, tt.expect)
			}
		})
	}
}

func TestGetMapInt64NilValue(t *testing.T) {
	tests := []struct {
		name   string
		header map[string]interface{}
		key    string
		expect *int64
	}{
		{"int value", map[string]interface{}{"a": int64(5)}, "a", ptrInt64(5)},
		{"float value", map[string]interface{}{"a": 5.7}, "a", ptrInt64(5)},
		{"uint value", map[string]interface{}{"a": uint64(7)}, "a", ptrInt64(7)},
		{"bool true", map[string]interface{}{"a": true}, "a", ptrInt64(1)},
		{"bool false", map[string]interface{}{"a": false}, "a", ptrInt64(0)},
		{"string int", map[string]interface{}{"a": "42"}, "a", ptrInt64(42)},
		{"string non-int", map[string]interface{}{"a": "abc"}, "a", nil},
		{"missing key", map[string]interface{}{"b": 1}, "a", nil},
		{"nil value", map[string]interface{}{"a": nil}, "a", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetMapInt64NilValue(tt.header, tt.key)
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("%s: got %v, want %v", tt.name, got, tt.expect)
			}
		})
	}
}

func ptrInt64(v int64) *int64 {
	return &v
}
