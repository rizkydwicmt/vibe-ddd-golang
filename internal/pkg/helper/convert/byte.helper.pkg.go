package convert

import "encoding/json"

func ByteToJSON(payload []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal(payload, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func ByteToStruct[I any](payload []byte) (result *I, err error) {
	err = json.Unmarshal(payload, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
