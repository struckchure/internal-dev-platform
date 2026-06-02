package ui

import (
	"encoding/json"
	"strings"
)

func formatWSStreamLine(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}

	var envelope struct {
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal([]byte(raw), &envelope); err != nil || len(envelope.Content) == 0 {
		return raw
	}

	if message := deploymentLogMessage(envelope.Content); message != "" {
		return message
	}

	return raw
}

func deploymentLogMessage(content json.RawMessage) string {
	var entry struct {
		Message *string `json:"message"`
	}
	if json.Unmarshal(content, &entry) == nil {
		if entry.Message != nil && strings.TrimSpace(*entry.Message) != "" {
			return strings.TrimSpace(*entry.Message)
		}
	}

	var nested string
	if json.Unmarshal(content, &nested) == nil && strings.TrimSpace(nested) != "" {
		if json.Unmarshal([]byte(nested), &entry) == nil {
			if entry.Message != nil && strings.TrimSpace(*entry.Message) != "" {
				return strings.TrimSpace(*entry.Message)
			}
		}
		return nested
	}

	return ""
}
