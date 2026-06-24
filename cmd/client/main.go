package main

import (
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

	headerStyle = lipgloss.NewStyle().
	Foreground(crust).
	Background(mauve).
	Bold(true).
	Padding(0, 1).
	MarginBottom(1)

	boxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(surface1).
	Padding(0, 1)

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

type model struct {
	state sessionState
	conn *websocket.Conn
	logs []string
	err error

	width int
	height int
}

func initModel() model {
	return model {
		state: stateConnecting,
		logs: []string{},
	}
}

func (m model) Init() tea.Cmd {
	return connectToServer
}

func connectToServer() tea.Msg {
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		return wsErr(fmt.Errorf("server offline: %v",err))
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
			case "1":
				m.logs = append(m.logs, cmdStyle.Render(">> Sent: Shadow Ball"))
				m.conn.WriteJSON(map[string]string{"act": "move", "value": "shadowball"})
			case "2":
				m.logs = append(m.logs, cmdStyle.Render(">> Sent: Tackle"))
				m.conn.WriteJSON(map[string]string{"act": "move", "value": "tackle"})
			}
		}

	case connectedMsg:
		m.conn = msg
		m.state = stateMatchmaking

		joinMsg := map[string]interface{}{
			"act":  "join",
			"team": []string{"gengar", "missingno", "charizardmegax"},
		}
		m.conn.WriteJSON(joinMsg)

		m.logs = append(m.logs, sysStyle.Render("[system] Connected, waiting for opps"))

		return m, listener(m.conn)

	case wsMsg:
		m.state = stateBattle
		m.logs = append(m.logs, string(msg))

		maxLogs := m.height - 10 
		if maxLogs < 5 { maxLogs = 5 }
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

	header := headerStyle.Width(m.width).Render(" 󰑔 POKEMON CLASH ENGINE ")

	contentWidth := m.width - 2
	contentHeight := m.height - 4
	var content string

	switch m.state {
	case stateConnecting:
		content = activeBoxStyle.Width(contentWidth).Height(contentHeight).
		Render(dimStyle.Render("\n Dialing server..."))

	case stateError:
		content = activeBoxStyle.Width(contentWidth).Height(contentHeight).
		Render(errStyle.Render(fmt.Sprintf("\n Connection Failed: %v\n\n", m.err)) +
		dimStyle.Render(" Make sure the Go engine is running."))

	case stateMatchmaking:
		var b strings.Builder
		b.WriteString(sysStyle.Render("\n MATCHMAKING\n\n"))
		for _, l := range m.logs {
			b.WriteString(" " + l + "\n")
		}
		content = activeBoxStyle.Width(contentWidth).Height(contentHeight).Render(b.String())

	case stateBattle:
		leftWidth := (contentWidth * 7) / 10
		var logs strings.Builder
		for _, l := range m.logs {
			logs.WriteString(l + "\n")
		}
		logPane := activeBoxStyle.Width(leftWidth).Height(contentHeight).Render(logs.String())

		rightWidth := contentWidth - leftWidth - 2
		var info strings.Builder
		
		info.WriteString(sysStyle.Render(" YOUR TEAM\n"))
		info.WriteString(" Gengar\n")
		info.WriteString(" Missingno\n")
		info.WriteString(" Charizard-Mega-X\n")
		
		info.WriteString("\n" + sysStyle.Render(" CONTROLS\n"))
		info.WriteString(dimStyle.Render(" 1 │") + " Shadow Ball\n")
		info.WriteString(dimStyle.Render(" 2 │") + " Tackle\n")
		info.WriteString(dimStyle.Render(" q │") + " Quit\n")
		
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
	p := tea.NewProgram(initModel(), tea.WithAltScreen()) 
	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}
}
