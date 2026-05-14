package utils

import "encoding/json"

// DecodeJSONMap decodes JSON bytes into a map[string]interface{}.
// Returns an empty map if the input is empty or invalid.
func DecodeJSONMap(raw []byte) map[string]interface{} {
	if len(raw) == 0 {
		return map[string]interface{}{}
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil || decoded == nil {
		return map[string]interface{}{}
	}
	return decoded
}
