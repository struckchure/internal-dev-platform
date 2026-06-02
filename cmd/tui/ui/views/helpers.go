package views

import "encoding/json"

func PrettyJSON(raw string) string {
	var val any
	if err := json.Unmarshal([]byte(raw), &val); err != nil {
		return raw
	}
	out, err := json.MarshalIndent(val, "", "  ")
	if err != nil {
		return raw
	}
	return string(out)
}
