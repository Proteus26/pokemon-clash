package engine

type Pokemon struct {
	Id string
	Mon string
	Level int
	Types []string
	Maxhp int
	Hp int 
	Atk int
	Def int
	Spa int
	Spd int
	Spe int
	//todo: add more stuff like abilities statuses moves and shit
}

type Player struct {
	Id string
	Team []*Pokemon
	Active *Pokemon
}

type Action struct {
	Pid string `json:"-"`
	Act string `json:"act"`
	Value string `json:"value,omitempty"`
	Team []string `json:"team,omitempty"`
}

type Battle struct {
	Id string
	P1 *Player
	P2 *Player

	P1chan chan Action
	P2chan chan Action
	Broadcast chan string
}
