package engine

import (
	"pokemon-clash/loader"
	"strings"
)

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

func (p *Player) Find(name string) *Pokemon {
	for _, mon := range p.Team {
		if strings.EqualFold(mon.Mon, name) || loader.ToID(mon.Mon) == loader.ToID(name) {
			return mon
		}
	}
	return nil
}

func (p *Player) Alive() bool {
	for _, mon := range p.Team {
		if mon.HP > 0 {
			return true
		}
	}
	return false
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

type Envelope struct {
	To   string
	Data []byte
}

type Battle struct {
	ID     string
	P1     *Player
	P2     *Player
	P1Chan chan Action
	P2Chan chan Action
	Out    chan Envelope
	Done   chan struct{}
	Turn   int
}

type ServerMessage struct {
	Type  string `json:"type"`
	Event string `json:"event,omitempty"`
	Text  string `json:"text"`

	P1Active string `json:"p1_active,omitempty"`
	P1HP     int    `json:"p1_hp"`
	P1Max    int    `json:"p1_max"`

	P2Active string `json:"p2_active,omitempty"`
	P2HP     int    `json:"p2_hp"`
	P2Max    int    `json:"p2_max"`

	Turn       int    `json:"turn,omitempty"`
	MustSwitch string `json:"must_switch,omitempty"`
	Winner     string `json:"winner,omitempty"`
}
