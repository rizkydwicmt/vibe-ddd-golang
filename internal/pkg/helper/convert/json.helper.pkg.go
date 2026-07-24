package convert

import "encoding/json"

// JSONToString marshals any value to its JSON string form.
func JSONToString(payload any) (string, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// JSONToByte marshals any value to its JSON byte form.
func JSONToByte(payload any) ([]byte, error) {
	return json.Marshal(payload)
}

// JSONToStruct round-trips any value through JSON into a typed *I. Useful for
// re-decoding loosely-typed transport values (e.g. an RPC response body) into
// a concrete struct.
func JSONToStruct[I any](payload any) (*I, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var result I
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MapToStruct decodes a map into a typed target via JSON.
func MapToStruct(data map[string]any, target any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}

// StructToMap encodes any value into a generic map via JSON.
func StructToMap(payload any) (map[string]any, error) {
	out := make(map[string]any)
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}
