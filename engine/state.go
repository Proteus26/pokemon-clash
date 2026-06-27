package engine

type Pokemon struct {
	ID    string
	Mon   string
	Level int
	Types []string
	Moves []string
	MaxHP int
	HP    int
	Atk   int
	Def   int
	SpA   int
	SpD   int
	Spe   int
}

type Player struct {
	ID     string
	Team   []*Pokemon
	Active *Pokemon
}

type TeamMember struct {
	Mon   string   `json:"mon"`
	Moves []string `json:"moves"`
}

type Action struct {
	PID   string       `json:"-"`
	Act   string       `json:"act"`
	Value string       `json:"value,omitempty"`
	Team  []TeamMember `json:"team,omitempty"`
}

type Battle struct {
	ID        string
	P1        *Player
	P2        *Player
	P1Chan    chan Action
	P2Chan    chan Action
	Broadcast chan []byte
}

type ServerMessage struct {
	Type  string `json:"type"`
	Event string `json:"event,omitempty"`
	Text  string `json:"text"`

	P1Active string `json:"p1_active,omitempty"`
	P1HP     int    `json:"p1_hp,omitempty"`
	P1Max    int    `json:"p1_max,omitempty"`

	P2Active string `json:"p2_active,omitempty"`
	P2HP     int    `json:"p2_hp,omitempty"`
	P2Max    int    `json:"p2_max,omitempty"`
}
