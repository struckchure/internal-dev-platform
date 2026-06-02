package ui

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
)

func (m *model) resetActionFormDefaults() {
	m.actionA = ""
	m.actionB = ""
	m.actionPayload = "{}"
	m.actionMachineID = ""
	m.actionRepoChoice = ""
	m.actionNetworkProtocol = "TCP"
	m.actionListeningPort = "8080"
	m.actionDestinationPort = "80"
	m.actionTitle = m.opsByTab[m.currentTab()][m.opIndex[m.currentTab()]]
}

func (m *model) applyActionFormToInputs() {
}

func (m *model) buildActionForm() {
	tab := m.currentTab()
	op := m.opIndex[tab]

	var fields []huh.Field
	switch tab {
	case "Auth+User":
		switch op {
		case 0:
			fields = append(fields,
				huh.NewInput().Title("Email").Value(&m.actionA),
				huh.NewInput().Title("Password").Value(&m.actionB).EchoMode(huh.EchoModePassword),
				huh.NewInput().Title("JSON payload").Description(`{"firstName":"...","lastName":"..."}`).Value(&m.actionPayload),
			)
		case 1:
			fields = append(fields,
				huh.NewInput().Title("Email").Value(&m.actionA),
				huh.NewInput().Title("Password").Value(&m.actionB).EchoMode(huh.EchoModePassword),
			)
		case 5:
			fields = append(fields,
				huh.NewInput().Title("Profile JSON payload").Description(`{"firstName":"..."}`).Value(&m.actionPayload),
			)
		}
	case "Machine+Network":
		switch op {
		case 2:
			fields = append(fields, m.machineSelectField("Machine", &m.actionA, false, ""))
		case 4:
			fields = append(fields, m.machineSelectField("Machine to delete", &m.actionA, false, ""))
		case 3:
			fields = append(fields,
				m.machineSelectField("Machine", &m.actionA, false, ""),
				huh.NewInput().Title("Update JSON payload").Value(&m.actionPayload),
			)
		case 6:
			fields = append(fields, m.machineSelectField("Machine", &m.actionMachineID, false, ""))
			fields = append(fields,
				huh.NewSelect[string]().
					Title("Protocol").
					Options(
						huh.NewOption("TCP", "TCP"),
						huh.NewOption("UDP", "UDP"),
					).
					Value(&m.actionNetworkProtocol),
				huh.NewInput().Title("Listening port").Value(&m.actionListeningPort),
				huh.NewInput().Title("Destination port").Value(&m.actionDestinationPort),
			)
		case 7:
			options := m.networkOptions()
			if len(options) > 0 {
				m.actionA = options[0].Value
				fields = append(fields, huh.NewSelect[string]().Title("Network to delete").Options(options...).Value(&m.actionA))
			} else {
				fields = append(fields, huh.NewInput().Title("Network ID").Value(&m.actionA))
			}
		}
	case "Repo+Deploy":
		switch op {
		case 1:
			fields = append(fields, m.machineSelectField("Machine", &m.actionMachineID, false, ""))
			repoOpts := m.repoOptions()
			if len(repoOpts) > 0 {
				m.actionRepoChoice = repoOpts[0].Value
				fields = append(fields, huh.NewSelect[string]().Title("Repository").Description("↑/↓ browse · tab next field").Options(repoOpts...).Value(&m.actionRepoChoice))
			} else {
				fields = append(fields,
					huh.NewInput().Title("GitHub repo ID").Value(&m.actionRepoChoice).Description("Connect GitHub first, or enter repo id"),
					huh.NewInput().Title("Repo name (owner/repo)").Value(&m.actionPayload),
				)
			}
		case 2, 4:
			connOpts := m.repoConnectionOptions()
			if len(connOpts) > 0 {
				m.actionA = connOpts[0].Value
				fields = append(fields, huh.NewSelect[string]().Title("Repo connection").Options(connOpts...).Value(&m.actionA))
			} else {
				fields = append(fields, huh.NewInput().Title("Connection ID").Value(&m.actionA))
			}
		case 3:
			connOpts := m.repoConnectionOptions()
			if len(connOpts) > 0 {
				m.actionA = connOpts[0].Value
				fields = append(fields, huh.NewSelect[string]().Title("Repo connection").Options(connOpts...).Value(&m.actionA))
			} else {
				fields = append(fields, huh.NewInput().Title("Connection ID").Value(&m.actionA))
			}
			m.actionMachineID = ""
			fields = append(fields, m.machineSelectField("Machine (optional)", &m.actionMachineID, true, "(unchanged)"))
			repoOpts := m.repoOptions()
			if len(repoOpts) > 0 {
				m.actionRepoChoice = ""
				fields = append(fields, huh.NewSelect[string]().Title("Repository (optional)").Description("↑/↓ browse · tab next field").Options(append([]huh.Option[string]{huh.NewOption("(unchanged)", "")}, repoOpts...)...).Value(&m.actionRepoChoice))
			}
		case 5:
			m.actionA = ""
			fields = append(fields, m.machineSelectField("Machine (optional)", &m.actionA, true, "(all machines)"))
		case 6:
			fields = append(fields, m.machineSelectField("Machine", &m.actionMachineID, false, ""))
			fields = append(fields, m.repoConnectionSelectField("Repo connection", &m.actionA, m.actionMachineID))
			fields = append(fields,
				huh.NewInput().
					Title("Git ref (optional)").
					Description("Branch or tag to deploy, e.g. main").
					Placeholder("main").
					Value(&m.actionB),
			)
		case 7, 8:
			fields = append(fields, huh.NewInput().Title("Deployment ID").Value(&m.actionA))
		}
	case "GitHub":
		if op == 0 {
			fields = append(fields,
				huh.NewInput().Title("Page number").Value(&m.actionA),
				huh.NewInput().Title("Page size").Value(&m.actionB),
			)
		}
	case "WebSocket":
		if op == 0 {
			fields = append(fields, huh.NewInput().Title("Event name").Value(&m.actionA).Placeholder(defaultWSEvent))
		}
	}

	group := huh.NewGroup(fields...).Title(m.actionTitle).WithShowHelp(true)
	m.actionForm = huh.NewForm(group).WithKeyMap(interactiveFormKeyMap())
}

const repoChoiceSep = "::"

func (m *model) parseRepoChoice() (repoID, repoName string) {
	choice := strings.TrimSpace(m.actionRepoChoice)
	if choice == "" {
		return "", ""
	}
	if parts := strings.SplitN(choice, repoChoiceSep, 2); len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	repoID = choice
	if name := strings.TrimSpace(m.actionPayload); name != "" {
		repoName = name
	}
	return repoID, repoName
}

func (m *model) repoConnectionPayloadFromForm() map[string]any {
	repoID, repoName := m.parseRepoChoice()
	return map[string]any{
		"machineId": strings.TrimSpace(m.actionMachineID),
		"repoId":    repoID,
		"repoName":  repoName,
	}
}

func (m *model) deployPayloadFromForm() map[string]any {
	return map[string]any{
		"connectionId": strings.TrimSpace(m.actionA),
		"ref":          strings.TrimSpace(m.actionB),
	}
}

func (m *model) repoConnectionUpdatePayloadFromForm() map[string]any {
	payload := map[string]any{}
	if id := strings.TrimSpace(m.actionMachineID); id != "" {
		payload["machineId"] = id
	}
	repoID, repoName := m.parseRepoChoice()
	if repoID != "" {
		payload["repoId"] = repoID
	}
	if repoName != "" {
		payload["repoName"] = repoName
	}
	return payload
}

func (m *model) networkPayloadFromForm() map[string]any {
	listeningPort, _ := strconv.Atoi(strings.TrimSpace(m.actionListeningPort))
	destinationPort, _ := strconv.Atoi(strings.TrimSpace(m.actionDestinationPort))

	return map[string]any{
		"machineId":       strings.TrimSpace(m.actionMachineID),
		"protocol":        strings.TrimSpace(m.actionNetworkProtocol),
		"listeningPort":   listeningPort,
		"destinationPort": destinationPort,
	}
}

func (m *model) machineNamesByID() map[string]string {
	items := m.machinesFromAPI()
	names := make(map[string]string, len(items))
	for _, item := range items {
		id := machineIDFromItem(item)
		if id == "" {
			continue
		}
		if name := machineNameFromItem(item); name != "" {
			names[id] = name
		} else {
			names[id] = id
		}
	}
	return names
}

func (m *model) machineLabel(machineID string, names map[string]string) string {
	machineID = strings.TrimSpace(machineID)
	if machineID == "" {
		return ""
	}
	if name, ok := names[machineID]; ok && name != "" && name != machineID {
		return name + " (" + machineID + ")"
	}
	return machineID
}

func (m *model) networkOptions() []huh.Option[string] {
	resp, status, err := m.api.ListNetworks()
	if err != nil || status >= 400 {
		return nil
	}

	var items []map[string]any
	if err := json.Unmarshal([]byte(resp), &items); err != nil {
		return nil
	}

	options := make([]huh.Option[string], 0, len(items))
	for _, item := range items {
		rawID, ok := item["id"]
		if !ok {
			continue
		}
		id, ok := rawID.(string)
		if !ok || id == "" {
			continue
		}

		label := id
		proto, _ := item["protocol"].(string)
		portVal, hasPort := item["listeningPort"]
		if proto != "" || hasPort {
			label = proto
			if hasPort {
				label = strings.TrimSpace(label + " " + toString(portVal))
			}
			label = strings.TrimSpace(label) + " (" + id + ")"
		}
		options = append(options, huh.NewOption(label, id))
	}
	return options
}

func (m *model) repoOptions() []huh.Option[string] {
	resp, status, err := m.api.ListRepos("1", "100")
	if err != nil || status >= 400 {
		return nil
	}

	var repos []struct {
		ID       float64 `json:"id"`
		Name     string  `json:"name"`
		FullName string  `json:"fullName"`
	}
	if err := json.Unmarshal([]byte(resp), &repos); err != nil {
		return nil
	}

	options := make([]huh.Option[string], 0, len(repos))
	for _, repo := range repos {
		if repo.ID == 0 {
			continue
		}
		repoID := strconv.Itoa(int(repo.ID))
		label := repo.FullName
		if label == "" {
			label = repo.Name
		}
		if label == "" {
			label = repoID
		} else {
			label = label + " (" + repoID + ")"
		}
		repoName := repo.FullName
		if repoName == "" {
			repoName = repo.Name
		}
		value := repoID + repoChoiceSep + repoName
		options = append(options, huh.NewOption(label, value))
	}
	return options
}

func repoConnectionMachineID(item map[string]any) string {
	if id := strings.TrimSpace(toString(item["machineId"])); id != "" {
		return id
	}
	if machine, ok := item["machine"].(map[string]any); ok {
		return machineIDFromItem(machine)
	}
	return ""
}

func (m *model) repoConnectionOptions() []huh.Option[string] {
	return m.repoConnectionOptionsForMachine("")
}

func (m *model) repoConnectionOptionsForMachine(machineID string) []huh.Option[string] {
	resp, status, err := m.api.ListRepoConnections()
	if err != nil || status >= 400 {
		return nil
	}

	var items []map[string]any
	if err := json.Unmarshal([]byte(resp), &items); err != nil {
		return nil
	}

	machineID = strings.TrimSpace(machineID)
	machineNames := m.machineNamesByID()
	options := make([]huh.Option[string], 0, len(items))
	for _, item := range items {
		connMachineID := repoConnectionMachineID(item)
		if machineID != "" && connMachineID != machineID {
			continue
		}

		id := strings.TrimSpace(toString(item["id"]))
		if id == "" {
			continue
		}
		label := id
		if name := strings.TrimSpace(toString(item["repoName"])); name != "" {
			label = name + " (" + id + ")"
		}
		if machineLabel := m.machineLabel(connMachineID, machineNames); machineLabel != "" && machineID == "" {
			label += " · " + machineLabel
		}
		options = append(options, huh.NewOption(label, id))
	}
	return options
}

func (m *model) repoConnectionSelectField(title string, value *string, machineID string) huh.Field {
	opts := m.repoConnectionOptionsForMachine(machineID)
	if len(opts) == 0 {
		opts = []huh.Option[string]{huh.NewOption("(no repo connections for this machine)", "")}
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

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return strconv.Itoa(int(x))
	case int:
		return strconv.Itoa(x)
	default:
		return ""
	}
}
