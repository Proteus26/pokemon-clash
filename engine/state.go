package engine

type Pokemon struct {
	Id string
	Mon string
	Maxhp int 
	Hp int 
	Spd int
	//todo: add more stuff like abilities statuses moves and shit
}

type Player struct {
	Id string
	Team []*Pokemon
	Active *Pokemon
}

type Action struct {
	Pid string
	Act string
	Value string
}

type Battle struct {
	Id string
	P1 *Player
	P2 *Player

	P1chan chan Action
	P2chan chan Action
	Broadcast chan string
}
