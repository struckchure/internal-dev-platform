package ui

import (
	"encoding/json"
	"strings"

	"github.com/charmbracelet/huh"
)

func (m *model) machinesFromAPI() []map[string]any {
	resp, status, err := m.api.ListMachines()
	if err != nil || status >= 400 {
		return nil
	}
	raw := []byte(resp)

	var items []map[string]any
	if err := json.Unmarshal(raw, &items); err == nil && len(items) > 0 {
		return items
	}

	var machines []struct {
		ID          string  `json:"id"`
		MachineName *string `json:"machineName"`
	}
	if err := json.Unmarshal(raw, &machines); err == nil && len(machines) > 0 {
		items = make([]map[string]any, 0, len(machines))
		for _, machine := range machines {
			if strings.TrimSpace(machine.ID) == "" {
				continue
			}
			item := map[string]any{"id": machine.ID}
			if machine.MachineName != nil {
				item["machineName"] = strings.TrimSpace(*machine.MachineName)
			}
			items = append(items, item)
		}
		return items
	}

	var wrapped struct {
		Data     []map[string]any `json:"data"`
		Items    []map[string]any `json:"items"`
		Machines []map[string]any `json:"machines"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil {
		for _, slice := range [][]map[string]any{wrapped.Data, wrapped.Items, wrapped.Machines} {
			if len(slice) > 0 {
				return slice
			}
		}
	}
	return nil
}

func machineIDFromItem(item map[string]any) string {
	return strings.TrimSpace(toString(item["id"]))
}

func machineNameFromItem(item map[string]any) string {
	if name := strings.TrimSpace(toString(item["machineName"])); name != "" {
		return name
	}
	return strings.TrimSpace(toString(item["name"]))
}

func (m *model) machineOptions() []huh.Option[string] {
	items := m.machinesFromAPI()
	options := make([]huh.Option[string], 0, len(items))
	for _, item := range items {
		id := machineIDFromItem(item)
		if id == "" {
			continue
		}
		label := id
		if name := machineNameFromItem(item); name != "" {
			label = name + " (" + id + ")"
		}
		options = append(options, huh.NewOption(label, id))
	}
	return options
}

func (m *model) machineSelectField(title string, value *string, optional bool, emptyLabel string) huh.Field {
	opts := m.machineOptions()
	if optional && emptyLabel != "" {
		opts = append([]huh.Option[string]{huh.NewOption(emptyLabel, "")}, opts...)
	}
	if len(opts) == 0 {
		opts = []huh.Option[string]{huh.NewOption("(no machines — create one on Machine+Network)", "")}
	}
	if strings.TrimSpace(*value) == "" {
		for _, opt := range opts {
			if strings.TrimSpace(opt.Value) != "" {
				*value = opt.Value
				break
			}
		}
	}
	return huh.NewSelect[string]().
		Title(title).
		Description("↑/↓ browse · tab next field").
		Options(opts...).
		Value(value)
}
