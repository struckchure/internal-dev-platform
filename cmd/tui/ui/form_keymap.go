package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
)

func interactiveFormKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()

	// Tab moves between fields; ↑/↓ stay on select options only.
	km.Input.Next = key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field"))
	km.Input.Prev = key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev field"))
	km.Input.Submit = key.NewBinding(key.WithKeys("enter", "ctrl+m"), key.WithHelp("enter", "submit"))

	km.Text.Next = key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field"))
	km.Text.Prev = key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev field"))
	km.Text.Submit = key.NewBinding(key.WithKeys("enter", "ctrl+m"), key.WithHelp("enter", "submit"))

	km.Select.Up = key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑", "up"))
	km.Select.Down = key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓", "down"))
	km.Select.Submit = key.NewBinding(key.WithKeys("enter", "ctrl+m"), key.WithHelp("enter", "submit"))
	km.Select.Next = key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field"))
	km.Select.Prev = key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev field"))

	km.MultiSelect.Up = key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑", "up"))
	km.MultiSelect.Down = key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓", "down"))
	km.MultiSelect.Submit = key.NewBinding(key.WithKeys("enter", "ctrl+m"), key.WithHelp("enter", "submit"))
	km.MultiSelect.Next = key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field"))
	km.MultiSelect.Prev = key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev field"))

	return km
}
