package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
)

var (
	text     = lipgloss.Color("#cdd6f4")
	base     = lipgloss.Color("#1e1e2e")
	crust    = lipgloss.Color("#11111b")
	mauve    = lipgloss.Color("#cba6f7")
	blue     = lipgloss.Color("#89b4fa")
	red      = lipgloss.Color("#f38ba8")
	green    = lipgloss.Color("#a6e3a1")
	surface1 = lipgloss.Color("#45475a")
	surface2 = lipgloss.Color("#585b70")

	headerStyle = lipgloss.NewStyle().Foreground(crust).Background(mauve).Bold(true)
	boxStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(surface1).Padding(0, 1)
	activeBoxStyle = boxStyle.BorderForeground(mauve)

	sysStyle = lipgloss.NewStyle().Foreground(blue).Bold(true)
	cmdStyle = lipgloss.NewStyle().Foreground(green)
	errStyle = lipgloss.NewStyle().Foreground(red).Bold(true)
	dimStyle = lipgloss.NewStyle().Foreground(surface2)
)

type wsMsg []byte
type wsErr error
type connectedMsg *websocket.Conn

type sessionState int

const (
	stateConnecting sessionState = iota
	stateMatchmaking
	stateBattle
	stateError
)

type ServerMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type LocalMon struct {
	Mon   string   `json:"mon"` 
	Moves []string `json:"moves"`
}

type model struct {
	state  sessionState
	conn   *websocket.Conn
	logs   []string
	err    error
	width  int
	height int
	team   []LocalMon
}

func initialModel(loadedTeam []LocalMon) model {
	return model{
		state: stateConnecting,
		logs:  []string{},
		team:  loadedTeam,
	}
}

func (m model) Init() tea.Cmd {
	return connectToServer
}

func connectToServer() tea.Msg {
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		return wsErr(fmt.Errorf("server offline: %v", err))
	}
	return connectedMsg(conn)
}

func listener(conn *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return wsErr(err)
		}
		return wsMsg(msg)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.conn != nil {
				m.conn.Close()
			}
			return m, tea.Quit
		}

		if m.state == stateBattle {
			switch msg.String() {
			case "1", "2", "3", "4":
				id := int(msg.String()[0] - '1')

				if id >= 0 && id < len(m.team[0].Moves) {
					selectedMove := m.team[0].Moves[id]
					m.logs = append(m.logs, cmdStyle.Render(">> Sent: "+selectedMove))
					m.conn.WriteJSON(map[string]string{"act": "move", "value": selectedMove})
				}
			}
		}

	case connectedMsg:
		m.conn = msg
		m.state = stateMatchmaking

		joinMsg := map[string]interface{}{
			"act":  "join",
			"team": m.team, 
		}
		m.conn.WriteJSON(joinMsg)
		m.logs = append(m.logs, sysStyle.Render("[System] Connected, waiting for opps"))
		return m, listener(m.conn)

	case wsMsg:
		m.state = stateBattle

		var sm ServerMessage
		if err := json.Unmarshal(msg, &sm); err == nil && sm.Text != "" {
			m.logs = append(m.logs, sm.Text)
		} else {
			m.logs = append(m.logs, string(msg))
		}

		maxLogs := m.height - 10
		if maxLogs < 5 {
			maxLogs = 5
		}
		if len(m.logs) > maxLogs {
			m.logs = m.logs[len(m.logs)-maxLogs:]
		}
		return m, listener(m.conn)

	case wsErr:
		m.state = stateError
		m.err = msg
		if m.conn != nil {
			m.conn.Close()
		}
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	headerText := " 󰨉 POKEMON CLASH ENGINE " 
	header := headerStyle.Width(m.width).Render(headerText)
	contentWidth := m.width - 2
	contentHeight := m.height - 4
	var content string

	switch m.state {
	case stateConnecting:
		content = activeBoxStyle.Width(contentWidth).Height(contentHeight).Render(dimStyle.Render("\n Dialing server..."))

	case stateError:
		content = activeBoxStyle.Width(contentWidth).Height(contentHeight).Render(errStyle.Render(fmt.Sprintf("\n Connection Failed: %v\n\n", m.err)) + dimStyle.Render(" Make sure the Go engine is running."))

	case stateMatchmaking:
		var b strings.Builder
		b.WriteString(sysStyle.Render("\n MATCHMAKING\n\n"))
		for _, l := range m.logs {
			b.WriteString(" ")
			b.WriteString(l)
			b.WriteString("\n")
		}
		content = activeBoxStyle.Width(contentWidth).Height(contentHeight).Render(b.String())

	case stateBattle:
		leftWidth := (contentWidth * 7) / 10
		var logs strings.Builder
		for _, l := range m.logs {
			logs.WriteString(l)
			logs.WriteString("\n")
		}
		logPane := activeBoxStyle.Width(leftWidth).Height(contentHeight).Render(logs.String())

		rightWidth := contentWidth - leftWidth - 2
		var info strings.Builder

		info.WriteString(sysStyle.Render(" YOUR TEAM\n"))
		for i, mon := range m.team {
			displayName := strings.ToUpper(string(mon.Mon[0])) + mon.Mon[1:]
			if i == 0 {
				info.WriteString(activeBoxStyle.Render(" ▶ " + displayName) + "\n")
			} else {
				info.WriteString("   " + displayName + "\n")
			}
		}

		info.WriteString("\n")
		info.WriteString(sysStyle.Render(" CONTROLS\n"))
		for i, move := range m.team[0].Moves {
			info.WriteString(dimStyle.Render(fmt.Sprintf(" %d │", i+1)))
			info.WriteString(" " + move + "\n")
		}
		info.WriteString(dimStyle.Render(" q │"))
		info.WriteString(" Quit\n")

		infoPane := boxStyle.Width(rightWidth).Height(contentHeight).Render(info.String())
		content = lipgloss.JoinHorizontal(lipgloss.Top, logPane, infoPane)
	}

	modeStr := " MATCHMAKING "
	modeBg := blue
	if m.state == stateBattle {
		modeStr = " BATTLE "
		modeBg = red
	} else if m.state == stateError {
		modeStr = " ERROR "
		modeBg = red
	}

	modeIndicator := lipgloss.NewStyle().Background(modeBg).Foreground(crust).Bold(true).Render(modeStr)
	connectionStatus := lipgloss.NewStyle().Background(surface1).Foreground(text).Padding(0, 1).Render("TCP :8080")

	statusBar := lipgloss.JoinHorizontal(lipgloss.Top, modeIndicator, connectionStatus)
	paddedStatus := lipgloss.NewStyle().Background(base).Width(m.width).Render(statusBar)

	return lipgloss.JoinVertical(lipgloss.Left, header, content, paddedStatus)
}

func main() {
	filename := "data/team1.json"
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	bytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("failed to load team file '%s': %v\n", filename, err)
		os.Exit(1)
	}

	var loadedTeam []LocalMon
	if err := json.Unmarshal(bytes, &loadedTeam); err != nil {
		fmt.Printf("Invalid JSON in team file: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel(loadedTeam), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
