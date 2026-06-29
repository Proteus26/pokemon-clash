package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
	"github.com/qeesung/image2ascii/convert"
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
	Type     string `json:"type"`
	Text     string `json:"text"`
	P1Active string `json:"p1_active"`
	P2Active string `json:"p2_active"`
	Role     string `json:"role,omitempty"`
}

type LocalMon struct {
	Mon   string   `json:"mon"` 
	Moves []string `json:"moves"`
}

type model struct {
	state       sessionState
	conn        *websocket.Conn
	logs        []string
	err         error
	width       int
	height      int
	team        []LocalMon
	p1Active    string
	p2Active    string
	spriteCache map[string]string
	myRole      string
}

func initialModel(loadedTeam []LocalMon) model {
	return model{
		state:       stateConnecting,
		logs:        []string{},
		team:        loadedTeam,
		spriteCache: make(map[string]string),
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
		if err := json.Unmarshal(msg, &sm); err == nil {
			if sm.Text != "" {
				m.logs = append(m.logs, sm.Text)
			}
			if sm.Role != "" {
				m.myRole = sm.Role
			}
			if m.myRole == "" {
				m.myRole = "p1"
			}

			var myMon, oppMon string
			if m.myRole == "p2" {
				myMon = sm.P2Active
				oppMon = sm.P1Active
			} else {
				myMon = sm.P1Active
				oppMon = sm.P2Active
			}

			converter := convert.NewImageConverter()
			opt :=  convert.DefaultOptions
			opt.FixedWidth = 35
			opt.FixedHeight = 18
			opt.Colored = true

			if myMon != "" {
				m.p1Active = strings.ToLower(myMon)
				if _, exists := m.spriteCache[m.p1Active+"_back"]; !exists {
					m.spriteCache[m.p1Active+"_back"] = converter.ImageFile2ASCIIString(fmt.Sprintf("data/sprites/back/%s.png", m.p1Active), &opt)
				}
			}

			if oppMon != "" {
				m.p2Active = strings.ToLower(oppMon)
				if _, exists := m.spriteCache[m.p2Active+"_front"]; !exists {
					m.spriteCache[m.p2Active+"_front"] = converter.ImageFile2ASCIIString(fmt.Sprintf("data/sprites/front/%s.png", m.p2Active), &opt)
				}
			}
		} else {
			m.logs = append(m.logs, string(msg))
		}

		maxLogs := m.height - 28
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
		topHeight := (contentHeight * 7) / 10
		bottomHeight := contentHeight - topHeight - 1
		leftWidth := (contentWidth * 6) / 10
		rightWidth := contentWidth - leftWidth - 2

		p1Sprite := m.spriteCache[m.p1Active+"_back"]
		p2Sprite := m.spriteCache[m.p2Active+"_front"]

		if p1Sprite == "" {
			p1Sprite = dimStyle.Render("\n Loading Sprite...")
		}
		if p2Sprite == "" {
			p2Sprite = dimStyle.Render("\n Loading Sprite...")
		}

		arenaStr := lipgloss.JoinHorizontal(lipgloss.Bottom,
		lipgloss.NewStyle().Width(leftWidth/2).Align(lipgloss.Left).Render(p1Sprite),
		lipgloss.NewStyle().Width(leftWidth/2).Align(lipgloss.Right).Render(p2Sprite),
	)
	arenaPane := activeBoxStyle.Width(leftWidth).Height(topHeight).Render(arenaStr)

	maxLogs := topHeight - 4
	if len(m.logs) > maxLogs {
		m.logs = m.logs[len(m.logs)-maxLogs:]
	}
	var logs strings.Builder
	logs.WriteString(sysStyle.Render(" BATTLE LOG\n\n"))
	for _, l := range m.logs {
		logs.WriteString(" " + l + "\n")
	}
	logPane := boxStyle.Width(rightWidth).Height(topHeight).Render(logs.String())

	var ctrls strings.Builder
	monName := "YOUR POKEMON"
	if m.p1Active != "" {
		monName = strings.ToUpper(m.p1Active)
	} else if len(m.team) > 0 {
		monName = strings.ToUpper(m.team[0].Mon)
	}
	ctrls.WriteString(sysStyle.Render(fmt.Sprintf(" WHAT WILL %s DO?\n\n", monName)))

	for i, move := range m.team[0].Moves {
		ctrls.WriteString(dimStyle.Render(fmt.Sprintf(" [%d] ", i+1)))
		ctrls.WriteString(lipgloss.NewStyle().Width(18).Render(move))
		if i == 1 {
			ctrls.WriteString("\n\n")
		}
	}
	ctrls.WriteString(dimStyle.Render("\n\n [q] ") + "Quit\n")
	controlsPane := activeBoxStyle.Width(leftWidth).Height(bottomHeight).Render(ctrls.String())

	var teamStr strings.Builder
	teamStr.WriteString(sysStyle.Render(" PARTY\n\n"))
	for i, mon := range m.team {
		displayName := strings.ToUpper(string(mon.Mon[0])) + mon.Mon[1:]
		if i == 0 {
			teamStr.WriteString(activeBoxStyle.Render(" ▶ " + displayName) + "\n")
		} else {
			teamStr.WriteString("   " + displayName + "\n")
		}
	}
	teamPane := boxStyle.Width(rightWidth).Height(bottomHeight).Render(teamStr.String())

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, arenaPane, logPane)
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, controlsPane, teamPane)
	content = lipgloss.JoinVertical(lipgloss.Left, topRow, bottomRow)
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
