package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"pokemon-clash/loader"

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
	yellow   = lipgloss.Color("#f9e2af")
	surface1 = lipgloss.Color("#45475a")
	surface2 = lipgloss.Color("#585b70")

	headerStyle    = lipgloss.NewStyle().Foreground(crust).Background(mauve).Bold(true)
	boxStyle       = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(surface1).Padding(0, 1)
	activeBoxStyle = boxStyle.BorderForeground(mauve)

	sysStyle    = lipgloss.NewStyle().Foreground(blue).Bold(true)
	cmdStyle    = lipgloss.NewStyle().Foreground(green)
	errStyle    = lipgloss.NewStyle().Foreground(red).Bold(true)
	dimStyle    = lipgloss.NewStyle().Foreground(surface2)
	activeStyle = lipgloss.NewStyle().Foreground(mauve).Bold(true)
	nameStyle   = lipgloss.NewStyle().Foreground(text).Bold(true)
)

var typeColors = map[string]lipgloss.Color{
	"Normal": "#a8a77a", "Fire": "#ee8130", "Water": "#6390f0", "Electric": "#f7d02c",
	"Grass": "#7ac74c", "Ice": "#96d9d6", "Fighting": "#c22e28", "Poison": "#a33ea1",
	"Ground": "#e2bf65", "Flying": "#a98ff3", "Psychic": "#f95587", "Bug": "#a6b91a",
	"Rock": "#b6a136", "Ghost": "#735797", "Dragon": "#6f35fc", "Dark": "#705746",
	"Steel": "#b7b7ce", "Fairy": "#d685ad",
}

const minWidth, minHeight = 80, 24

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
	Event    string `json:"event"`
	Text     string `json:"text"`
	P1Active string `json:"p1_active"`
	P1HP     int    `json:"p1_hp"`
	P1Max    int    `json:"p1_max"`
	P2Active string `json:"p2_active"`
	P2HP     int    `json:"p2_hp"`
	P2Max    int    `json:"p2_max"`
	Role     string `json:"role,omitempty"`
}

type LocalMon struct {
	Mon   string   `json:"mon"`
	Moves []string `json:"moves"`
}

type hp struct{ cur, max int }

type model struct {
	state     sessionState
	conn      *websocket.Conn
	logs      []string
	err       error
	width     int
	height    int
	team      []LocalMon
	myActive  string
	oppActive string
	myHP      hp
	oppHP     hp
	teamHP    map[string]hp
	sprites   *spriteStore
	myRole    string
	waiting   bool
	turn      int
}

const maxLogHistory = 100

var partyKeys = []string{"5", "6", "7", "8", "9", "0"}

func initialModel(loadedTeam []LocalMon) model {
	return model{
		state:   stateConnecting,
		logs:    []string{},
		team:    loadedTeam,
		teamHP:  map[string]hp{},
		sprites: newSpriteStore(),
		turn:    1,
	}
}

func (m model) activeIndex() int {
	for i, mon := range m.team {
		if strings.EqualFold(mon.Mon, m.myActive) {
			return i
		}
	}
	return 0
}

func (m model) fainted(mon string) bool {
	h, ok := m.teamHP[strings.ToLower(mon)]
	return ok && h.cur == 0
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

func (m *model) addLog(line string) {
	m.logs = append(m.logs, line)
	if len(m.logs) > maxLogHistory {
		m.logs = m.logs[len(m.logs)-maxLogHistory:]
	}
}

func (m *model) send(act, value, logLine string) {
	m.addLog(cmdStyle.Render(logLine))
	m.conn.WriteJSON(map[string]string{"act": act, "value": value})
	m.waiting = true
}

func (m model) handleBattleKey(key string) model {
	if m.waiting {
		return m
	}
	active := m.team[m.activeIndex()]

	switch key {
	case "1", "2", "3", "4":
		if m.fainted(active.Mon) {
			m.addLog(errStyle.Render(displayName(active.Mon) + " has fainted. Switch with [5-0]."))
			return m
		}
		id := int(key[0] - '1')
		if id < len(active.Moves) {
			move := active.Moves[id]
			m.send("move", move, "> "+displayName(active.Mon)+", use "+displayName(move)+"!")
		}

	case "5", "6", "7", "8", "9", "0":
		for idx, k := range partyKeys {
			if k != key || idx >= len(m.team) {
				continue
			}
			target := m.team[idx].Mon
			switch {
			case idx == m.activeIndex():
				m.addLog(dimStyle.Render(displayName(target) + " is already in battle."))
			case m.fainted(target):
				m.addLog(dimStyle.Render(displayName(target) + " has fainted and can't battle."))
			default:
				m.send("switch", target, "> Come back! Go, "+displayName(target)+"!")
			}
		}
	}
	return m
}

func (m model) handleServerMessage(raw []byte) model {
	var sm ServerMessage
	if err := json.Unmarshal(raw, &sm); err != nil {
		m.addLog(string(raw))
		return m
	}
	if sm.Role != "" {
		m.myRole = sm.Role
	}
	if m.myRole == "" {
		m.myRole = "p1"
	}

	if sm.Type == "battle" {
		m.state = stateBattle

		if m.waiting {
			m.waiting = false
			m.turn++
			m.addLog(dimStyle.Render(fmt.Sprintf("── Turn %d ──", m.turn-1)))
		}

		myMon, oppMon := sm.P1Active, sm.P2Active
		myHP, oppHP := hp{sm.P1HP, sm.P1Max}, hp{sm.P2HP, sm.P2Max}
		if m.myRole == "p2" {
			myMon, oppMon = oppMon, myMon
			myHP, oppHP = oppHP, myHP
		}
		if myMon != "" {
			m.myActive = strings.ToLower(myMon)
			m.myHP = myHP
			m.teamHP[m.myActive] = myHP
		}
		if oppMon != "" {
			m.oppActive = strings.ToLower(oppMon)
			m.oppHP = oppHP
		}
	}

	if sm.Text != "" {
		switch {
		case sm.Event == "faint":
			m.addLog(errStyle.Render(sm.Text))
		case sm.Event == "switch":
			m.addLog(sysStyle.Render(sm.Text))
		case sm.Type == "system":
			m.addLog(dimStyle.Render(sm.Text))
		default:
			m.addLog(sm.Text)
		}
	}
	return m
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
		if m.state == stateBattle && len(m.team) > 0 {
			m = m.handleBattleKey(msg.String())
		}

	case connectedMsg:
		m.conn = msg
		m.state = stateMatchmaking
		m.conn.WriteJSON(map[string]interface{}{"act": "join", "team": m.team})
		m.addLog(sysStyle.Render("Connected to server."))
		return m, listener(m.conn)

	case wsMsg:
		m = m.handleServerMessage(msg)
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

func displayName(id string) string {
	words := strings.Fields(strings.NewReplacer("_", " ", "-", " ").Replace(id))
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}

func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	return lipgloss.NewStyle().MaxWidth(w).Render(s)
}

func padRight(s string, w int) string {
	s = truncate(s, w)
	return s + strings.Repeat(" ", max(0, w-lipgloss.Width(s)))
}

func padLeft(s string, w int) string {
	s = truncate(s, w)
	return strings.Repeat(" ", max(0, w-lipgloss.Width(s))) + s
}

func innerSize(style lipgloss.Style, w, h int) (int, int) {
	return w - style.GetHorizontalFrameSize(), h - style.GetVerticalFrameSize()
}

func pane(style lipgloss.Style, w, h int, content string) string {
	innerW, innerH := innerSize(style, w, h)
	if innerW < 1 || innerH < 1 {
		return ""
	}
	lines := strings.Split(content, "\n")
	if len(lines) > innerH {
		lines = lines[:innerH]
	}
	for i, l := range lines {
		lines[i] = truncate(l, innerW)
	}
	return style.
		Width(w - style.GetHorizontalBorderSize()).
		Height(innerH).
		Render(strings.Join(lines, "\n"))
}

func hpBar(h hp, width int) string {
	if width < 1 {
		return ""
	}
	frac := 0.0
	if h.max > 0 {
		frac = float64(h.cur) / float64(h.max)
	}
	filled := int(frac*float64(width) + 0.5)
	if h.cur > 0 && filled == 0 {
		filled = 1
	}
	color := green
	if frac <= 0.2 {
		color = red
	} else if frac <= 0.5 {
		color = yellow
	}
	return dimStyle.Render("HP ") +
		lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("━", filled)) +
		lipgloss.NewStyle().Foreground(surface1).Render(strings.Repeat("─", width-filled))
}

func infoBox(name string, h hp, width int, showNumbers bool) []string {
	if name == "" {
		return nil
	}
	label := nameStyle.Render(displayName(name))
	if h.max > 0 && h.cur == 0 {
		label += " " + errStyle.Render("FNT")
	}
	lines := []string{label, hpBar(h, min(20, width-3))}
	if showNumbers && h.max > 0 {
		lines = append(lines, padLeft(dimStyle.Render(fmt.Sprintf("%d / %d", h.cur, h.max)), min(23, width)))
	}
	return lines
}

func place(rows [][2]string, side int, top int, block []string, w int, right bool) {
	for i, l := range block {
		if r := top + i; r >= 0 && r < len(rows) {
			if right {
				rows[r][side] = padLeft(l, w)
			} else {
				rows[r][side] = padRight(l, w)
			}
		}
	}
}

func spriteLines(s string) []string {
	if s == "" {
		return []string{dimStyle.Render("(no sprite)")}
	}
	return strings.Split(s, "\n")
}

func (m model) arenaView(w, h int) string {
	half := w / 2
	rows := make([][2]string, h)
	for i := range rows {
		rows[i] = [2]string{strings.Repeat(" ", half), strings.Repeat(" ", w-half)}
	}

	spriteH := max(4, h*3/5)
	oppSprite := spriteLines(m.sprites.get(fmt.Sprintf("data/sprites/front/%s.png", m.oppActive), w-half, spriteH))
	mySprite := spriteLines(m.sprites.get(fmt.Sprintf("data/sprites/back/%s.png", m.myActive), half, spriteH))

	place(rows, 1, 0, oppSprite, w-half, true)
	place(rows, 0, 1, infoBox(m.oppActive, m.oppHP, half-2, false), half, false)
	place(rows, 0, h-len(mySprite), mySprite, half, false)
	myInfo := infoBox(m.myActive, m.myHP, w-half, true)
	place(rows, 1, h-len(myInfo)-1, myInfo, w-half, true)

	out := make([]string, h)
	for i, r := range rows {
		out[i] = r[0] + r[1]
	}
	return strings.Join(out, "\n")
}

func (m model) logView(w, h int) string {
	wrap := lipgloss.NewStyle().Width(w)
	var lines []string
	for _, l := range m.logs {
		lines = append(lines, strings.Split(wrap.Render(l), "\n")...)
	}
	if avail := max(0, h-2); len(lines) > avail {
		lines = lines[len(lines)-avail:]
	}
	return sysStyle.Render("BATTLE LOG") + "\n\n" + strings.Join(lines, "\n")
}

func moveCell(i int, move string, nameW int, faded bool) string {
	name := padRight(displayName(move), nameW)
	if faded {
		return dimStyle.Render(fmt.Sprintf("[%d] ", i+1) + name)
	}
	cell := dimStyle.Render(fmt.Sprintf("[%d] ", i+1)) + name
	if data, ok := loader.GetMove(move); ok {
		badge := lipgloss.NewStyle().Foreground(crust).Background(typeColors[data.Type]).
			Render(" " + strings.ToUpper(data.Type) + " ")
		power := "—"
		if data.BP > 0 {
			power = fmt.Sprint(data.BP)
		}
		cell += " " + badge + dimStyle.Render(" "+power)
	}
	return cell
}

func (m model) movesView(w, h int) string {
	if len(m.team) == 0 {
		return dimStyle.Render("No team loaded.")
	}
	mon := m.team[m.activeIndex()]
	fainted := m.fainted(mon.Mon)

	var title string
	switch {
	case m.waiting:
		title = dimStyle.Render("Waiting for your opponent...")
	case fainted:
		title = errStyle.Render(fmt.Sprintf("%s FAINTED! CHOOSE A SWITCH [5-0]", strings.ToUpper(displayName(mon.Mon))))
	default:
		title = sysStyle.Render(fmt.Sprintf("WHAT WILL %s DO?", strings.ToUpper(displayName(mon.Mon))))
	}

	spaced := h >= 7
	lines := []string{title}
	if spaced {
		lines = append(lines, "")
	}

	nameW := 0
	for _, mv := range mon.Moves {
		nameW = max(nameW, lipgloss.Width(displayName(mv)))
	}
	nameW = min(nameW, max(1, w/2-5))
	for i := 0; i < len(mon.Moves) && i < 4; i += 2 {
		row := padRight(moveCell(i, mon.Moves[i], nameW, m.waiting || fainted), w/2)
		if i+1 < len(mon.Moves) {
			row += moveCell(i+1, mon.Moves[i+1], nameW, m.waiting || fainted)
		}
		lines = append(lines, row)
		if spaced {
			lines = append(lines, "")
		}
	}
	lines = append(lines, dimStyle.Render("[1-4]")+" Move   "+dimStyle.Render("[5-0]")+" Switch   "+dimStyle.Render("[q]")+" Quit")
	return strings.Join(lines, "\n")
}

func (m model) partyView(w int) string {
	lines := []string{sysStyle.Render("PARTY"), ""}
	active := m.activeIndex()
	for i, mon := range m.team {
		if i >= len(partyKeys) {
			break
		}
		key := "[" + partyKeys[i] + "]"
		status := ""
		if h, ok := m.teamHP[strings.ToLower(mon.Mon)]; ok {
			if h.cur == 0 {
				status = errStyle.Render("FNT")
			} else {
				status = dimStyle.Render(fmt.Sprintf("%d/%d", h.cur, h.max))
			}
		}

		var left string
		switch {
		case i == active:
			left = activeStyle.Render("▶ " + key + " " + displayName(mon.Mon))
		case m.fainted(mon.Mon):
			left = dimStyle.Render("  " + key + " " + displayName(mon.Mon))
		default:
			left = "  " + dimStyle.Render(key) + " " + displayName(mon.Mon)
		}
		gap := w - lipgloss.Width(left) - lipgloss.Width(status)
		lines = append(lines, left+strings.Repeat(" ", max(1, gap))+status)
	}
	return strings.Join(lines, "\n")
}

func (m model) battleView(w, h int) string {
	topH := h * 7 / 10
	botH := h - topH
	leftW := w * 6 / 10
	rightW := w - leftW

	arenaW, arenaH := innerSize(activeBoxStyle, leftW, topH)
	logW, logH := innerSize(boxStyle, rightW, topH)
	movesW, movesH := innerSize(activeBoxStyle, leftW, botH)
	partyW, _ := innerSize(boxStyle, rightW, botH)

	top := lipgloss.JoinHorizontal(lipgloss.Top,
		pane(activeBoxStyle, leftW, topH, m.arenaView(arenaW, arenaH)),
		pane(boxStyle, rightW, topH, m.logView(logW, logH)),
	)
	bottom := lipgloss.JoinHorizontal(lipgloss.Top,
		pane(activeBoxStyle, leftW, botH, m.movesView(movesW, movesH)),
		pane(boxStyle, rightW, botH, m.partyView(partyW)),
	)
	return lipgloss.JoinVertical(lipgloss.Left, top, bottom)
}

func (m model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	header := headerStyle.Width(m.width).Render(" 󰨉 POKEMON CLASH ENGINE ")
	contentHeight := m.height - 2
	var content string

	switch {
	case m.state == stateBattle && (m.width < minWidth || m.height < minHeight):
		msg := errStyle.Render("Terminal too small") + "\n\n" +
			dimStyle.Render(fmt.Sprintf("Resize to at least %dx%d (now %dx%d).", minWidth, minHeight, m.width, m.height))
		content = lipgloss.Place(m.width, contentHeight, lipgloss.Center, lipgloss.Center, msg)

	case m.state == stateConnecting:
		content = pane(activeBoxStyle, m.width, contentHeight, dimStyle.Render("Dialing server..."))

	case m.state == stateError:
		content = pane(activeBoxStyle, m.width, contentHeight,
			errStyle.Render(fmt.Sprintf("Connection Failed: %v", m.err))+"\n\n"+
				dimStyle.Render("Make sure the Go engine is running."))

	case m.state == stateMatchmaking:
		lines := append([]string{sysStyle.Render("MATCHMAKING"), ""}, m.logs...)
		content = pane(activeBoxStyle, m.width, contentHeight, strings.Join(lines, "\n"))

	case m.state == stateBattle:
		content = m.battleView(m.width, contentHeight)
	}

	modeStr := " MATCHMAKING "
	modeBg := blue
	if m.state == stateBattle {
		modeStr = fmt.Sprintf(" BATTLE · TURN %d ", m.turn)
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

	log.SetOutput(io.Discard)
	if err := loader.LoadAll(); err != nil {
		fmt.Printf("warning: move info unavailable (%v)\n", err)
	}

	p := tea.NewProgram(initialModel(loadedTeam), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
