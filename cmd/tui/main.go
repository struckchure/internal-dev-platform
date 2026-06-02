package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/struckchure/idp/cmd/tui/client"
	"github.com/struckchure/idp/cmd/tui/ui"
)

func main() {
	httpClient := client.NewAPIClient("http://localhost:3000")
	wsClient := client.NewWSClient("ws://localhost:9090/ws")

	program := tea.NewProgram(
		ui.NewModel(httpClient, wsClient),
		tea.WithAltScreen(),
	)

	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}
