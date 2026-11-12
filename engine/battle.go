package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"pokemon-clash/loader"
	"time"
)

const (
	DefaultTurnTimeout = 30 * time.Second
	DefaultMaxTimeouts = 3
)

func InitBattle(id string, p1, p2 *Player) *Battle {
	return &Battle{
		ID:     id,
		P1:     p1,
		P2:     p2,
		P1Chan: make(chan Action),
		P2Chan: make(chan Action),
		Out:    make(chan Envelope, 100),
		Done:   make(chan struct{}),
		Turn:   1,
		leave:  make(chan *Player),

		TurnTimeout: DefaultTurnTimeout,
		MaxTimeouts: DefaultMaxTimeouts,
		timeouts:    map[*Player]int{},
	}
}

func (b *Battle) other(p *Player) *Player {
	if p == b.P1 {
		return b.P2
	}
	return b.P1
}

func (b *Battle) role(p *Player) string {
	if p == b.P1 {
		return "p1"
	}
	return "p2"
}

func (b *Battle) label(p *Player) string {
	if p == b.P1 {
		return "Player 1"
	}
	return "Player 2"
}

func (b *Battle) emit(to string, msg ServerMessage) {
	msg.Turn = b.Turn
	bytes, _ := json.Marshal(msg)
	b.Out <- Envelope{To: to, Data: bytes}
}

func (b *Battle) snapshot(event, text string) ServerMessage {
	return ServerMessage{
		Type:     "battle",
		Event:    event,
		Text:     text,
		P1Active: b.P1.Active.Mon,
		P1HP:     b.P1.Active.HP,
		P1Max:    b.P1.Active.MaxHP,
		P2Active: b.P2.Active.Mon,
		P2HP:     b.P2.Active.HP,
		P2Max:    b.P2.Active.MaxHP,
	}
}

func (b *Battle) EmitSystem(text string) {
	b.emit("", ServerMessage{Type: "system", Text: text})
}

func (b *Battle) EmitBattle(event, text string) {
	b.emit("", b.snapshot(event, text))
}

func (b *Battle) emitError(p *Player, text string) {
	b.emit(b.role(p), ServerMessage{Type: "error", Text: text})
}

func (b *Battle) needsSwitch(p *Player) bool {
	return p.Active.HP == 0 && p.Alive()
}

func (b *Battle) forcedPhase() bool {
	return b.needsSwitch(b.P1) || b.needsSwitch(b.P2)
}

func (b *Battle) mustSwitch() string {
	p1, p2 := b.needsSwitch(b.P1), b.needsSwitch(b.P2)
	switch {
	case p1 && p2:
		return "both"
	case p1:
		return "p1"
	case p2:
		return "p2"
	}
	return ""
}

func (b *Battle) request() {
	b.seq++
	msg := b.snapshot("request", "")
	msg.MustSwitch = b.mustSwitch()
	msg.Seq = b.seq
	msg.TimeLimit = int(b.TurnTimeout.Round(time.Second) / time.Second)
	for _, p := range []*Player{b.P1, b.P2} {
		if b.needsSwitch(p) {
			msg.Text = fmt.Sprintf("%s must send out another Pokémon!", b.label(p))
			if msg.MustSwitch == "both" {
				msg.Text = "Both players must send out another Pokémon!"
			}
		}
	}
	b.emit("", msg)
}

func (b *Battle) validate(p *Player, act Action) error {
	switch act.Act {
	case "move":
		if p.Active.HP == 0 {
			return fmt.Errorf("%s has fainted. Choose a replacement.", p.Active.Mon)
		}
		for _, m := range p.Active.Moves {
			if loader.ToID(m) == loader.ToID(act.Value) {
				return nil
			}
		}
		return fmt.Errorf("%s doesn't know %s.", p.Active.Mon, act.Value)
	case "switch":
		target := p.Find(act.Value)
		switch {
		case target == nil:
			return fmt.Errorf("You don't have a Pokémon named %s.", act.Value)
		case target == p.Active:
			return fmt.Errorf("%s is already in battle.", target.Mon)
		case target.HP == 0:
			return fmt.Errorf("%s has fainted and can't battle.", target.Mon)
		}
		return nil
	}
	return errors.New("Unknown action.")
}

func (b *Battle) submit(p *Player, act Action, pending map[*Player]*Action) {
	switch {
	case act.Seq != b.seq:
		b.emitError(p, "Too late! That choice was for an earlier turn.")
		return
	case b.forcedPhase() && !b.needsSwitch(p):
		b.emitError(p, "Waiting for your opponent to send out a Pokémon.")
		return
	case b.forcedPhase() && act.Act != "switch":
		b.emitError(p, fmt.Sprintf("%s has fainted. Choose a replacement.", p.Active.Mon))
		return
	case pending[p] != nil:
		b.emitError(p, "You've already chosen an action this turn.")
		return
	}
	if err := b.validate(p, act); err != nil {
		b.emitError(p, err.Error())
		return
	}
	pending[p] = &act
	b.timeouts[p] = 0
}

func (b *Battle) required(p *Player) bool {
	return !b.forcedPhase() || b.needsSwitch(p)
}

func (b *Battle) autoAction(p *Player) *Action {
	if b.forcedPhase() {
		var bench []*Pokemon
		for _, mon := range p.Team {
			if mon.HP > 0 && mon != p.Active {
				bench = append(bench, mon)
			}
		}
		if len(bench) == 0 {
			return nil
		}
		return &Action{Act: "switch", Value: bench[rand.Intn(len(bench))].Mon, Seq: b.seq}
	}
	if len(p.Active.Moves) == 0 {
		return nil
	}
	return &Action{Act: "move", Value: p.Active.Moves[rand.Intn(len(p.Active.Moves))], Seq: b.seq}
}

func (b *Battle) handleTimeout(pending map[*Player]*Action) bool {
	for _, p := range []*Player{b.P1, b.P2} {
		if !b.required(p) || pending[p] != nil {
			continue
		}
		b.timeouts[p]++
		if b.timeouts[p] >= b.MaxTimeouts {
			b.finish(b.role(b.other(p)), fmt.Sprintf("%s ran out of time %d turns in a row and forfeits! %s wins the battle!",
				b.label(p), b.timeouts[p], b.label(b.other(p))))
			return true
		}
		act := b.autoAction(p)
		if act == nil {
			continue
		}
		pending[p] = act
		choice := act.Value
		if md, ok := loader.GetMove(act.Value); ok && act.Act == "move" {
			choice = md.Name
		}
		left := b.MaxTimeouts - b.timeouts[p]
		b.EmitSystem(fmt.Sprintf("%s ran out of time! %s was chosen for them. (%d more and they forfeit)",
			b.label(p), choice, left))
	}
	return false
}

func (b *Battle) ready(pending map[*Player]*Action) bool {
	for _, p := range []*Player{b.P1, b.P2} {
		if b.required(p) && pending[p] == nil {
			return false
		}
	}
	return true
}

func (b *Battle) resolve(pending map[*Player]*Action) {
	if b.forcedPhase() {
		for _, p := range []*Player{b.P1, b.P2} {
			if act := pending[p]; act != nil {
				b.switchIn(p, act.Value)
			}
		}
		return
	}
	b.moveOrder(*pending[b.P1], *pending[b.P2])
	b.Turn++
}

func (b *Battle) winner() string {
	p1, p2 := b.P1.Alive(), b.P2.Alive()
	switch {
	case !p1 && !p2:
		return "draw"
	case !p2:
		return "p1"
	case !p1:
		return "p2"
	}
	return ""
}

func (b *Battle) finish(winner, text string) {
	if text == "" {
		text = "It's a draw!"
		switch winner {
		case "p1":
			text = "Player 1 wins the battle!"
		case "p2":
			text = "Player 2 wins the battle!"
		}
	}
	msg := b.snapshot("end", text)
	msg.Winner = winner
	b.emit("", msg)
	log.Printf("[Engine] Battle %s over: %s\n", b.ID, text)
}

func (b *Battle) Leave(p *Player) {
	select {
	case b.leave <- p:
	case <-b.Done:
	}
}

func (b *Battle) Start() {
	defer close(b.Done)
	defer close(b.Out)

	log.Printf("[Engine] Battle %s initialized.\n", b.ID)
	b.EmitBattle("start", "The battle has begun!")
	b.request()

	turnTimer := time.NewTimer(b.TurnTimeout)
	defer turnTimer.Stop()

	pending := map[*Player]*Action{}

	for {
		select {
		case act := <-b.P1Chan:
			b.submit(b.P1, act, pending)
		case act := <-b.P2Chan:
			b.submit(b.P2, act, pending)
		case p := <-b.leave:
			b.finish(b.role(b.other(p)), fmt.Sprintf("%s disconnected. %s wins the battle!", b.label(p), b.label(b.other(p))))
			return
		case <-turnTimer.C:
			if b.handleTimeout(pending) {
				return
			}
			turnTimer.Reset(b.TurnTimeout)
		}

		if !b.ready(pending) {
			continue
		}

		b.resolve(pending)
		clear(pending)

		if w := b.winner(); w != "" {
			b.finish(w, "")
			return
		}

		b.request()
		turnTimer.Reset(b.TurnTimeout)
	}
}
