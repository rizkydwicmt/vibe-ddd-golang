package database

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONRawValue(t *testing.T) {
	tests := []struct {
		name    string
		input   JSONRaw
		want    driver.Value
		wantErr bool
	}{
		{name: "nil", input: nil, want: nil},
		{name: "empty", input: JSONRaw(""), want: nil},
		{name: "object", input: JSONRaw(`{"ok":true}`), want: `{"ok":true}`},
		{name: "array", input: JSONRaw(`[1,2]`), want: `[1,2]`},
		{name: "invalid", input: JSONRaw(`{"broken"`), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.Value()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestJSONRawScan(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    JSONRaw
		wantErr bool
	}{
		{name: "nil", input: nil, want: nil},
		{name: "bytes", input: []byte(`{"a":1}`), want: JSONRaw(`{"a":1}`)},
		{name: "string", input: `{"a":2}`, want: JSONRaw(`{"a":2}`)},
		{name: "raw message", input: json.RawMessage(`{"a":3}`), want: JSONRaw(`{"a":3}`)},
		{name: "empty bytes", input: []byte{}, want: nil},
		{name: "empty string", input: "", want: nil},
		{name: "unsupported", input: 1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got JSONRaw
			err := got.Scan(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestJSONRawDBDataType(t *testing.T) {
	tests := map[string]string{
		"":          "JSON",
		"postgres":  "JSONB",
		"mysql":     "JSON",
		"sqlite":    "JSON",
		"sqlite3":   "JSON",
		"sqlserver": "NVARCHAR(MAX)",
		"oracle":    "JSON",
	}

	for dialect, want := range tests {
		t.Run(dialect, func(t *testing.T) {
			assert.Equal(t, want, jsonRawDBDataType(dialect))
		})
	}
}

func TestJSONRawMarshalUnmarshalAndCompact(t *testing.T) {
	got, err := JSONRaw(nil).MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, []byte("null"), got)

	var raw JSONRaw
	input := []byte(`{"b": 2, "a": 1}`)
	require.NoError(t, raw.UnmarshalJSON(input))
	input[1] = 'x'
	assert.Equal(t, JSONRaw(`{"b": 2, "a": 1}`), raw)
	assert.Equal(t, JSONRaw(`{"b":2,"a":1}`), raw.Compact())
	assert.Equal(t, JSONRaw(`{"broken"`), JSONRaw(`{"broken"`).Compact())
	assert.False(t, raw.IsEmpty())
	assert.True(t, JSONRaw(nil).IsEmpty())
}
