package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
)

var (
	text = lipgloss.Color("#cdd6f4")
	mauve = lipgloss.Color("#cba6f7")
	blue = lipgloss.Color("#89b4fa")
	red = lipgloss.Color("#f38ba8")
	green = lipgloss.Color("#a6e3a1")
	surface1 = lipgloss.Color("#45475a")

	appStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(mauve).
		Padding(1, 2).
		Foreground(text)

	sysStyle = lipgloss.NewStyle().Foreground(blue).Bold(true)
	cmdStyle = lipgloss.NewStyle().Foreground(green)
	errStyle = lipgloss.NewStyle().Foreground(red).Bold(true)
	dimStyle = lipgloss.NewStyle().Foreground(surface1)
)

type wsmsg []byte
type wserr error

type model struct {
	conn *websocket.Conn
	logs []string
	err error
}

func initmodel() model {
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatal("Failed to connect to server: ", err)
	}

	joinmsg := map[string]interface{}{
		"act": "join",
		"team": []string{"gengar", "missingno", "charizardmegax"},
	}
	conn.WriteJSON(joinmsg)

	return model{
		conn: conn,
		logs: []string{sysStyle.Render("Connected to engine adn awaiting matchmaking")},
	}
}

func (m model) Init() tea.Cmd {
	return listener(m.conn)
}

func listener(conn *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return wserr(err)
		}
		return wsmsg(msg)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.conn.Close()
			return m, tea.Quit
		case "1":
			m.logs = append(m.logs, cmdStyle.Render("Sent: Shadow Ball"))
			m.conn.WriteJSON(map[string]string{"act": "move", "value": "shadowball"})
			return m, nil
		case "2":
			m.logs = append(m.logs, cmdStyle.Render("Sent: Tackle"))
			m.conn.WriteJSON(map[string]string{"act": "move", "value": "tackle"})
			return m, nil
		}

	case wsmsg:
		m.logs = append(m.logs, string(msg))
		
		if len(m.logs) > 15 {
			m.logs = m.logs[1:] 
		}
		return m, listener(m.conn)
		
	case wserr:
		m.err = msg
		m.logs = append(m.logs, errStyle.Render("Connection lost"))
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return errStyle.Render(fmt.Sprintf("Error: %v\nPress 'q' to quit.", m.err))
	}

	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Foreground(mauve).Bold(true).Render("POKEMON CLASH ENGINE\n"))

	for _, l := range m.logs {
		b.WriteString(l + "\n")
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("----------------------------------\n"))
	b.WriteString(" Controls: [1] Shadow Ball  [2] Tackle  [q] Quit\n")

	return appStyle.Render(b.String())
}

func main() {
	p := tea.NewProgram(initmodel(), tea.WithAltScreen()) 
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
