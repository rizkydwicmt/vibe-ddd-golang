// Package db: JSONRaw is the entity field type for JSON columns whose
// payloads the service stores and forwards without decoding at the
// persistence boundary. It mirrors `encoding/json.RawMessage`, but also
// satisfies `database/sql.Scanner` and `driver.Valuer` so it works with
// drivers that return JSON as []byte, string, or json.RawMessage.
//
// Use this type in entities whose JSON column is opaque transport JSON
// (e.g. raw provider payloads, normalized envelopes, settings, metadata).
// Use typed value objects when the column has a fixed schema the
// application must validate.
package database

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// JSONRaw mirrors `encoding/json.RawMessage` semantics: it stores raw
// JSON bytes verbatim and defers any structured decoding to the caller.
// Construct with `db.JSONRaw(b)` from a `[]byte` slice, just like
// `json.RawMessage(b)`. The zero value is `nil`, meaning "no JSON" —
// callers that need an explicit empty-object should use
// `db.JSONRaw("{}")`.
type JSONRaw []byte

// MarshalJSON returns the raw bytes unchanged. A nil/empty receiver is
// emitted as the JSON literal `null` so encoders that don't honor
// `omitempty` still produce valid JSON. This matches `json.RawMessage`.
func (j JSONRaw) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

// UnmarshalJSON stores the input bytes verbatim. Mirrors
// `json.RawMessage.UnmarshalJSON`: the caller-supplied JSON is preserved
// byte-for-byte so downstream re-marshaling produces identical output.
func (j *JSONRaw) UnmarshalJSON(b []byte) error {
	if j == nil {
		return errors.New("db.JSONRaw: UnmarshalJSON on nil pointer")
	}
	// Copy so callers can mutate the input slice safely; this also
	// avoids holding a reference into a decoder's internal buffer.
	*j = append((*j)[:0], b...)
	return nil
}

// Value implements driver.Valuer. Empty/nil maps to SQL NULL so the column
// default (`'{}'` on tables that set one) kicks in rather than inserting an
// invalid empty string into a JSON column.
func (j JSONRaw) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil //nolint:nilnil // driver.Valuer contract: nil maps to SQL NULL
	}
	// Validate so we never persist a non-JSON byte sequence.
	if !json.Valid(j) {
		return nil, errors.New("db.JSONRaw: value is not valid JSON")
	}
	// string is accepted by JSON, TEXT, and NVARCHAR-backed JSON columns.
	return string(j), nil
}

// Scan implements sql.Scanner. Accepts the three shapes the underlying
// driver may return for a JSON column:
//
//	nil             → zero the receiver (NULL column).
//	[]byte          → store verbatim.
//	string          → coerce to []byte.
//	json.RawMessage → copy into the receiver.
func (j *JSONRaw) Scan(value any) error {
	if j == nil {
		return errors.New("db.JSONRaw: Scan into nil receiver")
	}
	switch v := value.(type) {
	case nil:
		*j = nil
		return nil
	case []byte:
		// Copy: drivers may reuse their buffer between rows.
		if len(v) == 0 {
			*j = nil
			return nil
		}
		buf := make([]byte, len(v))
		copy(buf, v)
		*j = buf
		return nil
	case string:
		if len(v) == 0 {
			*j = nil
			return nil
		}
		*j = []byte(v)
		return nil
	case json.RawMessage:
		if len(v) == 0 {
			*j = nil
			return nil
		}
		buf := make([]byte, len(v))
		copy(buf, v)
		*j = buf
		return nil
	default:
		return fmt.Errorf("db.JSONRaw: unsupported scan type %T", value)
	}
}

// GormDataType pins the logical cross-dialect type to `json`.
func (JSONRaw) GormDataType() string { return "json" }

// GormDBDataType returns the dialect-specific column type.
func (JSONRaw) GormDBDataType(tx *gorm.DB, _ *schema.Field) string {
	if tx == nil {
		return jsonRawDBDataType("")
	}
	return jsonRawDBDataType(tx.Dialector.Name())
}

func jsonRawDBDataType(dialect string) string {
	switch dialect {
	case "postgres":
		return "JSONB"
	case "mysql", "sqlite", "sqlite3":
		return "JSON"
	case "sqlserver":
		return "NVARCHAR(MAX)"
	default:
		return "JSON"
	}
}

// String returns the raw JSON as a string — useful for logging and
// debugging. NOT redaction-safe; do not pass to user-visible surfaces
// without going through internal/pkg/redaction first.
func (j JSONRaw) String() string {
	if len(j) == 0 {
		return ""
	}
	return string(j)
}

// IsEmpty reports whether the receiver carries no payload (either nil
// or zero-length). Callers preferring slice-length idioms can keep
// using `len(j) == 0`; this method exists for readability in business
// code that doesn't otherwise reach into the byte layer.
func (j JSONRaw) IsEmpty() bool { return len(j) == 0 }

// Compact strips insignificant whitespace from the stored JSON. Useful
// before equality checks since GORM may round-trip JSONB with reordered
// whitespace. Best-effort: returns the original bytes when compaction
// fails (e.g. invalid JSON).
func (j JSONRaw) Compact() JSONRaw {
	if len(j) == 0 {
		return nil
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, j); err != nil {
		return j
	}
	return JSONRaw(buf.Bytes())
}
