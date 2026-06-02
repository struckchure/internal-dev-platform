package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/struckchure/idp/cmd/tui/client"
)

type keyMap struct {
	NextSection key.Binding
	PrevSection key.Binding
	NextInput   key.Binding
	PrevInput   key.Binding
	Run         key.Binding
	SubmitForm  key.Binding
	OpenLink    key.Binding
	Help        key.Binding
	Clear       key.Binding
	Quit        key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.PrevSection, k.NextSection, k.NextInput, k.Run, k.SubmitForm, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.PrevSection, k.NextSection, k.NextInput, k.PrevInput},
		{k.Run, k.SubmitForm, k.OpenLink, k.Help, k.Clear, k.Quit},
	}
}

type wsMsg string

type apiResultMsg struct {
	title         string
	status        int
	body          string
	err           error
	authenticated bool
	autoWS        bool
}

type wsSubscribedMsg struct {
	event     string
	err       error
	connected bool
}

type model struct {
	api *client.APIClient
	ws  *client.WSClient

	tabs     []string
	tabIndex int

	inputs     []textinput.Model
	focusIndex int

	opsByTab map[string][]string
	opIndex  map[string]int

	response viewport.Model
	status   string
	loading  bool
	spin     spinner.Model

	wsEvent     string
	wsMessages  []string
	wsSink      chan string
	wsConnected bool
	showHelp    bool

	showMachineForm bool
	machineForm     *huh.Form
	machineName     string
	machineCPU      string
	machineMemory   string
	machineImage    string

	showActionForm        bool
	actionForm            *huh.Form
	actionTitle           string
	actionA               string
	actionB               string
	actionPayload         string
	actionMachineID       string
	actionRepoChoice      string
	actionNetworkProtocol string
	actionListeningPort   string
	actionDestinationPort string
	lastOpenURL           string

	keys keyMap
	help help.Model

	width  int
	height int
}

func NewModel(api *client.APIClient, ws *client.WSClient) tea.Model {
	baseHTTP := textinput.New()
	baseHTTP.Placeholder = "HTTP base URL"
	baseHTTP.SetValue("http://localhost:3000")
	baseHTTP.Width = 56

	baseWS := textinput.New()
	baseWS.Placeholder = "WS base URL"
	baseWS.SetValue("ws://localhost:9090/ws")
	baseWS.Width = 56

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = focusStyle

	vp := viewport.New(80, 20)
	vp.SetContent(usageGuide)

	return &model{
		api:      api,
		ws:       ws,
		tabs:     []string{"Auth+User", "Machine+Network", "Repo+Deploy", "GitHub", "WebSocket"},
		inputs:   []textinput.Model{baseHTTP, baseWS},
		opsByTab: operationsByTab(),
		opIndex: map[string]int{
			"Auth+User":       0,
			"Machine+Network": 0,
			"Repo+Deploy":     0,
			"GitHub":          0,
			"WebSocket":       0,
		},
		response: vp,
		status:   "Ready — press ? for help",
		spin:     s,
		wsSink:   make(chan string, 100),
		showHelp: true,
		keys: keyMap{
			NextInput:   key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next input")),
			PrevInput:   key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev input")),
			NextSection: key.NewBinding(key.WithKeys("shift+right"), key.WithHelp("shift+→", "next tab")),
			PrevSection: key.NewBinding(key.WithKeys("shift+left"), key.WithHelp("shift+←", "prev tab")),
			Run:         key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "run action")),
			SubmitForm:  key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "submit form")),
			OpenLink:    key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open last link")),
			Help:        key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle guide")),
			Clear:       key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "clear output")),
			Quit:        key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		},
		help: help.New(),
	}
}

func operationsByTab() map[string][]string {
	return map[string][]string{
		"Auth+User": {
			"Register",
			"Login",
			"Refresh Access Token",
			"Logout",
			"Get Profile",
			"Update Profile",
		},
		"Machine+Network": {
			"List Machines",
			"Create Machine (interactive form)",
			"Get Machine",
			"Update Machine",
			"Delete Machine",
			"List Networks",
			"Create Network",
			"Delete Network",
		},
		"Repo+Deploy": {
			"List Repo Connections",
			"Create Repo Connection",
			"Get Repo Connection",
			"Update Repo Connection",
			"Delete Repo Connection",
			"List Deployments",
			"Deploy Repo",
			"Get Deployment",
			"List Deployment Logs",
		},
		"GitHub": {
			"List Repositories",
			"Authorize Account Link",
			"Update App Access Link",
			"List Account Connections",
		},
		"WebSocket": {
			"Subscribe",
			"Disconnect",
		},
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spin.Tick, m.waitForWS())
}

func (m *model) waitForWS() tea.Cmd {
	return func() tea.Msg {
		return wsMsg(<-m.wsSink)
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.response.Width = max(20, msg.Width-8)
		m.response.Height = max(8, msg.Height-17)
	case tea.KeyMsg:
		if m.showMachineForm {
			if key.Matches(msg, m.keys.Quit) || msg.String() == "esc" {
				m.showMachineForm = false
				m.status = "Machine creation cancelled"
				return m, nil
			}
			if msg.String() == "enter" || msg.String() == "ctrl+m" || msg.String() == "ctrl+j" || msg.String() == "ctrl+s" {
				m.showMachineForm = false
				m.applyBases()
				return m, m.runCreateMachine(map[string]any{
					"machineName":  strings.TrimSpace(m.machineName),
					"cpu":          strings.TrimSpace(m.machineCPU),
					"memory":       strings.TrimSpace(m.machineMemory),
					"machineImage": strings.TrimSpace(m.machineImage),
				})
			}
			if msg.String() == "enter" || msg.String() == "ctrl+m" {
				form, cmd := m.machineForm.Update(huh.NextField())
				if f, ok := form.(*huh.Form); ok {
					m.machineForm = f
				}
				switch m.machineForm.State {
				case huh.StateCompleted:
					m.showMachineForm = false
					m.applyBases()
					return m, m.runCreateMachine(map[string]any{
						"machineName":  strings.TrimSpace(m.machineName),
						"cpu":          strings.TrimSpace(m.machineCPU),
						"memory":       strings.TrimSpace(m.machineMemory),
						"machineImage": strings.TrimSpace(m.machineImage),
					})
				case huh.StateAborted:
					m.showMachineForm = false
					m.status = "Machine creation cancelled"
					return m, cmd
				}
				return m, cmd
			}
			if msg.String() == "ctrl+s" {
				for i := 0; i < 24 && m.machineForm.State == huh.StateNormal; i++ {
					form, _ := m.machineForm.Update(huh.NextField())
					if f, ok := form.(*huh.Form); ok {
						m.machineForm = f
					}
				}
				if m.machineForm.State == huh.StateCompleted {
					m.showMachineForm = false
					m.applyBases()
					return m, m.runCreateMachine(map[string]any{
						"machineName":  strings.TrimSpace(m.machineName),
						"cpu":          strings.TrimSpace(m.machineCPU),
						"memory":       strings.TrimSpace(m.machineMemory),
						"machineImage": strings.TrimSpace(m.machineImage),
					})
				}
				return m, nil
			}
			if msg.String() == "tab" || msg.String() == "shift+tab" {
				var form tea.Model
				var cmd tea.Cmd
				if msg.String() == "shift+tab" {
					form, cmd = m.machineForm.Update(huh.PrevField())
				} else {
					form, cmd = m.machineForm.Update(huh.NextField())
				}
				if f, ok := form.(*huh.Form); ok {
					m.machineForm = f
				}
				return m, cmd
			}
			form, cmd := m.machineForm.Update(msg)
			if f, ok := form.(*huh.Form); ok {
				m.machineForm = f
			}
			switch m.machineForm.State {
			case huh.StateCompleted:
				m.showMachineForm = false
				m.applyBases()
				return m, m.runCreateMachine(map[string]any{
					"machineName":  strings.TrimSpace(m.machineName),
					"cpu":          strings.TrimSpace(m.machineCPU),
					"memory":       strings.TrimSpace(m.machineMemory),
					"machineImage": strings.TrimSpace(m.machineImage),
				})
			case huh.StateAborted:
				m.showMachineForm = false
				m.status = "Machine creation cancelled"
				return m, cmd
			}
			return m, cmd
		}
		if m.showActionForm {
			if key.Matches(msg, m.keys.Quit) || msg.String() == "esc" {
				m.showActionForm = false
				m.status = "Action cancelled"
				return m, nil
			}
			if msg.String() == "enter" || msg.String() == "ctrl+m" || msg.String() == "ctrl+j" || msg.String() == "ctrl+s" {
				m.showActionForm = false
				m.applyActionFormToInputs()
				m.applyBases()
				return m, m.runCurrentOperation()
			}
			if msg.String() == "enter" || msg.String() == "ctrl+m" {
				form, cmd := m.actionForm.Update(huh.NextField())
				if f, ok := form.(*huh.Form); ok {
					m.actionForm = f
				}
				switch m.actionForm.State {
				case huh.StateCompleted:
					m.showActionForm = false
					m.applyActionFormToInputs()
					m.applyBases()
					return m, m.runCurrentOperation()
				case huh.StateAborted:
					m.showActionForm = false
					m.status = "Action cancelled"
					return m, cmd
				}
				return m, cmd
			}
			if msg.String() == "ctrl+s" {
				for i := 0; i < 24 && m.actionForm.State == huh.StateNormal; i++ {
					form, _ := m.actionForm.Update(huh.NextField())
					if f, ok := form.(*huh.Form); ok {
						m.actionForm = f
					}
				}
				if m.actionForm.State == huh.StateCompleted {
					m.showActionForm = false
					m.applyActionFormToInputs()
					m.applyBases()
					return m, m.runCurrentOperation()
				}
				return m, nil
			}
			if msg.String() == "tab" || msg.String() == "shift+tab" {
				var form tea.Model
				var cmd tea.Cmd
				if msg.String() == "shift+tab" {
					form, cmd = m.actionForm.Update(huh.PrevField())
				} else {
					form, cmd = m.actionForm.Update(huh.NextField())
				}
				if f, ok := form.(*huh.Form); ok {
					m.actionForm = f
				}
				return m, cmd
			}
			form, cmd := m.actionForm.Update(msg)
			if f, ok := form.(*huh.Form); ok {
				m.actionForm = f
			}
			switch m.actionForm.State {
			case huh.StateCompleted:
				m.showActionForm = false
				m.applyActionFormToInputs()
				m.applyBases()
				return m, m.runCurrentOperation()
			case huh.StateAborted:
				m.showActionForm = false
				m.status = "Action cancelled"
				return m, cmd
			}
			return m, cmd
		}
		if key.Matches(msg, m.keys.Quit) {
			_ = m.ws.Close()
			return m, tea.Quit
		}
		if key.Matches(msg, m.keys.NextSection) {
			m.tabIndex = (m.tabIndex + 1) % len(m.tabs)
			m.onTabChanged()
		}
		if key.Matches(msg, m.keys.PrevSection) {
			m.tabIndex = (m.tabIndex - 1 + len(m.tabs)) % len(m.tabs)
			m.onTabChanged()
		}
		if key.Matches(msg, m.keys.NextInput) {
			m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
		}
		if key.Matches(msg, m.keys.PrevInput) {
			m.focusIndex = (m.focusIndex - 1 + len(m.inputs)) % len(m.inputs)
		}
		if key.Matches(msg, m.keys.Help) {
			m.showHelp = !m.showHelp
			if m.showHelp {
				m.response.SetContent(usageGuide)
				m.response.GotoTop()
			}
		}
		if key.Matches(msg, m.keys.OpenLink) {
			if m.lastOpenURL == "" {
				m.status = mutedStyle.Render("No link available yet")
			} else if err := openURL(m.lastOpenURL); err != nil {
				m.status = errorStyle.Render("Open link failed: " + err.Error())
			} else {
				m.status = successStyle.Render("Opened link: " + m.lastOpenURL)
			}
		}
		if key.Matches(msg, m.keys.Clear) {
			m.response.SetContent("")
			m.wsMessages = nil
			m.status = "Cleared"
			m.showHelp = false
		}
		if msg.String() == "up" {
			current := m.currentTab()
			ops := m.opsByTab[current]
			m.opIndex[current] = (m.opIndex[current] - 1 + len(ops)) % len(ops)
		}
		if msg.String() == "down" {
			current := m.currentTab()
			ops := m.opsByTab[current]
			m.opIndex[current] = (m.opIndex[current] + 1) % len(ops)
		}
		if key.Matches(msg, m.keys.Run) && !m.loading {
			m.showHelp = false
			m.applyBases()
			if m.wantsMachineForm() {
				m.resetMachineFormDefaults()
				m.buildMachineForm()
				m.showMachineForm = true
				return m, m.machineForm.Init()
			}
			if m.wantsActionForm() {
				m.resetActionFormDefaults()
				m.buildActionForm()
				m.showActionForm = true
				return m, m.actionForm.Init()
			}
			return m, m.runCurrentOperation()
		}
	case spinner.TickMsg:
		m.spin, cmd = m.spin.Update(msg)
		cmds = append(cmds, cmd)
	case apiResultMsg:
		m.loading = false
		m.status = m.formatStatus(msg)
		m.captureOpenURL(msg.body)
		if msg.authenticated {
			m.showHelp = false
		}
		if strings.Contains(msg.body, "logged out") {
			m.wsConnected = false
			m.wsEvent = ""
			m.wsMessages = nil
		}
		m.response.SetContent(msg.body)
		m.response.GotoTop()
		if msg.autoWS && m.api.IsAuthenticated() {
			cmds = append(cmds, m.subscribeWS(defaultWSEvent))
		}
	case wsSubscribedMsg:
		m.loading = false
		if msg.err != nil {
			m.wsConnected = false
			m.wsEvent = ""
			m.status = errorStyle.Render("WebSocket: " + msg.err.Error())
		} else if msg.connected {
			m.wsConnected = true
			m.wsEvent = msg.event
			m.status = successStyle.Render("WebSocket subscribed: " + msg.event)
		} else {
			m.wsConnected = false
			m.wsEvent = ""
		}
	case wsMsg:
		line := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), string(msg))
		m.wsMessages = append(m.wsMessages, line)
		if len(m.wsMessages) > 300 {
			m.wsMessages = m.wsMessages[len(m.wsMessages)-300:]
		}
		if !m.showHelp {
			m.renderStreamOutput()
		}
		cmds = append(cmds, m.waitForWS())
	}

	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	m.response, cmd = m.response.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) currentTab() string {
	return m.tabs[m.tabIndex]
}

func (m *model) onTabChanged() {
}

func (m *model) wantsMachineForm() bool {
	return m.currentTab() == "Machine+Network" && m.opIndex["Machine+Network"] == 1
}

func (m *model) wantsActionForm() bool {
	tab := m.currentTab()
	op := m.opIndex[tab]

	switch tab {
	case "Auth+User":
		return op == 0 || op == 1 || op == 5
	case "Machine+Network":
		return op == 2 || op == 3 || op == 4 || op == 6 || op == 7
	case "Repo+Deploy":
		return op >= 1 && op <= 8
	case "GitHub":
		return op == 0
	case "WebSocket":
		return op == 0
	default:
		return false
	}
}

func (m *model) runCreateMachine(payload map[string]any) tea.Cmd {
	m.loading = true
	m.status = "Creating machine..."
	api := m.api
	return func() tea.Msg {
		body, status, err := api.CreateMachine(payload)
		return apiResultMsg{title: "Create Machine", status: status, body: body, err: err}
	}
}

func (m *model) applyBases() {
	m.api.SetBaseURL(strings.TrimSpace(m.inputs[0].Value()))
	m.ws.SetBaseURL(strings.TrimSpace(m.inputs[1].Value()))
}

func (m *model) parsePayload() (map[string]any, error) {
	payloadText := strings.TrimSpace(m.actionPayload)
	if payloadText == "" {
		return map[string]any{}, nil
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(payloadText), &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON payload: %w", err)
	}
	return payload, nil
}

func (m *model) formatStatus(msg apiResultMsg) string {
	auth := mutedStyle.Render("○ not authenticated")
	if m.api.IsAuthenticated() {
		auth = successStyle.Render("● authenticated (" + m.api.TokenPreview() + ")")
	}

	if msg.err != nil {
		return errorStyle.Render("Error: "+msg.err.Error()) + "  " + auth
	}
	if msg.status >= 400 {
		return errorStyle.Render(fmt.Sprintf("%s (%d)", msg.title, msg.status)) + "  " + auth
	}
	line := successStyle.Render(fmt.Sprintf("%s (%d)", msg.title, msg.status))
	if msg.authenticated {
		line += "  " + successStyle.Render("token saved — used on all API calls")
	}
	if m.lastOpenURL != "" {
		line += "  " + mutedStyle.Render("press o to open link")
	}
	return line + "  " + auth
}

func (m *model) captureOpenURL(body string) {
	var payload any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return
	}
	m.lastOpenURL = findURL(payload)
}

func findURL(value any) string {
	switch v := value.(type) {
	case map[string]any:
		for _, key := range []string{"link", "url", "htmlUrl", "htmlURL"} {
			if raw, ok := v[key]; ok {
				if s, ok := raw.(string); ok && isHTTPURL(s) {
					return s
				}
			}
		}
		for _, child := range v {
			if found := findURL(child); found != "" {
				return found
			}
		}
	case []any:
		for _, item := range v {
			if found := findURL(item); found != "" {
				return found
			}
		}
	}
	return ""
}

func isHTTPURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func openURL(raw string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", raw)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", raw)
	default:
		cmd = exec.Command("xdg-open", raw)
	}
	return cmd.Start()
}

func (m *model) renderStreamOutput() {
	if len(m.wsMessages) == 0 {
		return
	}
	m.response.SetContent(strings.Join(m.wsMessages, "\n"))
	m.response.GotoBottom()
}

func (m *model) subscribeWS(event string) tea.Cmd {
	event = strings.TrimSpace(event)
	if event == "" {
		event = defaultWSEvent
	}
	ws := m.ws
	sink := m.wsSink
	return func() tea.Msg {
		err := ws.Subscribe(event, sink)
		return wsSubscribedMsg{event: event, err: err, connected: err == nil}
	}
}

func (m *model) runCurrentOperation() tea.Cmd {
	m.loading = true
	m.status = "Running..."
	tab := m.currentTab()
	op := m.opsByTab[tab][m.opIndex[tab]]
	a := strings.TrimSpace(m.actionA)
	b := strings.TrimSpace(m.actionB)

	payload, payloadErr := m.parsePayload()
	api := m.api
	ws := m.ws
	wsSink := m.wsSink
	opIdx := m.opIndex[tab]

	return func() tea.Msg {
		if payloadErr != nil {
			return apiResultMsg{title: op, status: 400, body: payloadErr.Error(), err: payloadErr}
		}
		var body string
		var status int
		var err error
		var authenticated bool
		var autoWS bool

		switch tab {
		case "Auth+User":
			switch opIdx {
			case 0:
				firstName, _ := payload["firstName"].(string)
				lastName, _ := payload["lastName"].(string)
				body, status, err = api.Register(firstName, lastName, a, b)
				authenticated = api.IsAuthenticated() && status < 400
				autoWS = authenticated
			case 1:
				body, status, err = api.Login(a, b)
				authenticated = api.IsAuthenticated() && status < 400
				autoWS = authenticated
			case 2:
				body, status, err = api.Refresh()
				authenticated = api.IsAuthenticated() && status < 400
				autoWS = authenticated
			case 3:
				api.Logout()
				_ = ws.Close()
				body, status, err = `{"status":"logged out"}`, http.StatusOK, nil
			case 4:
				body, status, err = api.GetProfile()
			case 5:
				body, status, err = api.UpdateProfile(payload)
			}
		case "Machine+Network":
			switch opIdx {
			case 0:
				body, status, err = api.ListMachines()
			case 2:
				body, status, err = api.GetMachine(a)
			case 3:
				body, status, err = api.UpdateMachine(a, payload)
			case 4:
				body, status, err = api.DeleteMachine(a)
			case 5:
				body, status, err = api.ListNetworks()
			case 6:
				body, status, err = api.CreateNetwork(m.networkPayloadFromForm())
			case 7:
				body, status, err = api.DeleteNetwork(a)
			}
		case "Repo+Deploy":
			switch opIdx {
			case 0:
				body, status, err = api.ListRepoConnections()
			case 1:
				body, status, err = api.CreateRepoConnection(m.repoConnectionPayloadFromForm())
			case 2:
				body, status, err = api.GetRepoConnection(a)
			case 3:
				body, status, err = api.UpdateRepoConnection(a, m.repoConnectionUpdatePayloadFromForm())
			case 4:
				body, status, err = api.DeleteRepoConnection(a)
			case 5:
				body, status, err = api.ListDeployments(a)
			case 6:
				body, status, err = api.DeployRepo(m.deployPayloadFromForm())
			case 7:
				body, status, err = api.GetDeployment(a)
			case 8:
				body, status, err = api.ListDeploymentLogs(a)
			}
		case "GitHub":
			switch opIdx {
			case 0:
				body, status, err = api.ListRepos(a, b)
			case 1:
				body, status, err = api.AuthorizeGithub()
			case 2:
				body, status, err = api.UpdateAppAccess()
			case 3:
				body, status, err = api.ListAccountConnections()
			}
		case "WebSocket":
			switch opIdx {
			case 0:
				event := a
				if event == "" {
					event = defaultWSEvent
				}
				err = ws.Subscribe(event, wsSink)
				if err == nil {
					body, status = fmt.Sprintf(`{"status":"subscribed","event":"%s"}`, event), http.StatusOK
					return wsSubscribedMsg{event: event, err: nil, connected: true}
				}
				return wsSubscribedMsg{event: event, err: err, connected: false}
			case 1:
				err = ws.Close()
				body, status = `{"status":"disconnected"}`, http.StatusOK
				return wsSubscribedMsg{event: "", err: err, connected: false}
			}
		}

		return apiResultMsg{
			title:         op,
			status:        status,
			body:          body,
			err:           err,
			authenticated: authenticated,
			autoWS:        autoWS,
		}
	}
}

func (m *model) View() string {
	if m.showMachineForm && m.machineForm != nil {
		return appPadding.Render(lipgloss.JoinVertical(
			lipgloss.Left,
			titleStyle.Render("idp — Create machine"),
			mutedStyle.Render("Fill each field, then confirm. Esc or q cancels."),
			m.machineForm.View(),
		))
	}
	if m.showActionForm && m.actionForm != nil {
		return appPadding.Render(lipgloss.JoinVertical(
			lipgloss.Left,
			titleStyle.Render("idp — "+m.actionTitle),
			mutedStyle.Render("Fill fields, then confirm. Esc or q cancels."),
			m.actionForm.View(),
		))
	}

	tabs := make([]string, 0, len(m.tabs))
	for i, t := range m.tabs {
		if i == m.tabIndex {
			tabs = append(tabs, tabActiveStyle.Render(t))
			continue
		}
		tabs = append(tabs, tabStyle.Render(t))
	}

	currentTab := m.currentTab()
	ops := m.opsByTab[currentTab]
	opRows := make([]string, 0, len(ops))
	for i, op := range ops {
		prefix := "  "
		rowStyle := mutedStyle
		if i == m.opIndex[currentTab] {
			prefix = "▶ "
			rowStyle = focusStyle
		}
		opRows = append(opRows, rowStyle.Render(prefix+op))
	}

	inputRows := []string{
		m.labelInput("HTTP", m.inputs[0], 0),
		m.labelInput("WS", m.inputs[1], 1),
	}

	left := panelStyle.Width(max(40, m.width/2-6)).Render(strings.Join(append([]string{
		titleStyle.Render("Actions"),
		strings.Join(opRows, "\n"),
		"",
		titleStyle.Render("Config"),
	}, inputRows...), "\n"))

	header := titleStyle.Render("idp API Monitor")
	if m.loading {
		header += " " + m.spin.View()
	}

	authInfo := mutedStyle.Render("Auth: not logged in")
	if m.api.IsAuthenticated() {
		authInfo = successStyle.Render("Auth: logged in · token " + m.api.TokenPreview())
	}

	wsInfo := mutedStyle.Render("WS: disconnected")
	if m.wsConnected {
		wsInfo = successStyle.Render("WS: streaming " + m.wsEvent)
	}

	guideHint := mutedStyle.Render("Press ? guide · o open-link · shift+←/→ sections · ↑/↓ actions · tab/shift+tab form fields")

	rightContent := strings.Join([]string{
		header,
		authInfo,
		wsInfo,
		guideHint,
		"",
		m.response.View(),
	}, "\n")

	right := panelStyle.Width(max(40, m.width/2-6)).Render(rightContent)

	statusBar := panelStyle.Width(max(20, m.width-8)).Render(m.status)
	helpBar := mutedStyle.Render(m.help.View(m.keys))

	return appPadding.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		strings.Join(tabs, " "),
		lipgloss.JoinHorizontal(lipgloss.Top, left, right),
		statusBar,
		helpBar,
	))
}

func (m *model) labelInput(label string, input textinput.Model, idx int) string {
	lbl := mutedStyle.Render(label + ": ")
	if idx == m.focusIndex {
		lbl = focusStyle.Render(label + ": ")
	}
	return lbl + input.View()
}
