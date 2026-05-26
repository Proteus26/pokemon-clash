package engine

type pokemon struct {
	id string
	mon string
	maxhp int 
	hp int 
	spd int
	//todo: add more stuff like abilities statuses moves and shit
}

type player struct {
	id string
	team []*pokemon
	active *pokemon
}

type action struct {
	pid string
	act string
	value string
}

type battle struct {
	id string
	p1 *player
	p2 *player

	p1chan chan action
	p2chan chan action
	broadcast chan string
}
