// Package sim plays Eat the King between bots, with no server, browser or
// database, so thousands of games can be played to see how the game behaves.
package sim

import (
	"math/rand/v2"

	"github.com/streamlinedesign/chess-bots/engine"
)

// View is everything a player knows when it's their move: what a real player
// gets from the server, and nothing more. Under fog, squares the player can't
// see are empty in Board and false in Seen.
type View struct {
	Me    engine.Color
	Ply   int
	Fog   bool
	Board [64]engine.Piece
	Seen  [64]bool
	// Legal is the moves the server would offer. Every square a move lands on
	// is one the player sees, so the list leaks nothing the board doesn't.
	Legal []engine.Move
	// Castling is the player's own castling rights (the opponent's are unknown).
	Castling uint8
	Rules    engine.Rules
	// Danger is set when any enemy piece attacks Me's king, seen or not: the
	// site tells a player that much, never where the attacker is.
	Danger bool
	// Sense is the rule idea in play (see vision), so a bot can play to it.
	Sense string
	// pos is the real position, for bots such as the site's own (package bots)
	// that are handed it with the seen mask and hide the fog themselves.
	pos engine.Position
}

// Bot picks moves for one side of one game. A new Bot is made for every game,
// so a bot may remember what it saw on earlier turns.
type Bot interface {
	Move(v *View) engine.Move
}

// BotMaker makes a bot for one game, seeded so a game can be played again exactly.
type BotMaker func(rng *rand.Rand) Bot

// Bots are the players the simulator knows, by name.
var Bots = map[string]BotMaker{
	"random": func(rng *rand.Rand) Bot { return &randomBot{rng} },
	"greedy": func(rng *rand.Rand) Bot { return &greedyBot{rng: rng} },
	"hunter": func(rng *rand.Rand) Bot { return &hunter{rng: rng} },
	"wary":   func(rng *rand.Rand) Bot { return &hunter{rng: rng, wary: true} },
	// unaware is wary that doesn't know the sense in play and so never plays
	// for it, though it still sees what the sense shows it.
	"unaware": func(rng *rand.Rand) Bot { return &hunter{rng: rng, wary: true, unaware: true} },
}

// viewOf builds what color c knows in pos.
func viewOf(g *engine.Game, c engine.Color, fog bool, sense string, rules engine.Rules) *View {
	pos := g.Position()
	v := &View{Me: c, Ply: len(g.History()), Fog: fog, Sense: sense, Legal: g.LegalMoves(), Rules: rules, pos: pos}
	if fog {
		v.Seen = vision(&pos, c, sense, lastArrival(g, c))
	} else {
		for i := range v.Seen {
			v.Seen[i] = true
		}
	}
	for i, pc := range pos.Board {
		if v.Seen[i] {
			v.Board[i] = pc
		}
	}
	mine := engine.CastleWK | engine.CastleWQ
	if c == engine.Black {
		mine = engine.CastleBK | engine.CastleBQ
	}
	v.Castling = pos.Castling & mine
	if k := pos.KingSquare(c); k != engine.NoSquare {
		v.Danger = pos.Attacked(k, c.Other())
	}
	return v
}

// lastArrival is the square c's most recent move landed on, or NoSquare.
func lastArrival(g *engine.Game, c engine.Color) engine.Square {
	h := g.History()
	for i := len(h) - 1; i >= 0 && i >= len(h)-2; i-- {
		if pc := g.Position().Board[h[i].Move.To]; pc.Is(c) {
			return h[i].Move.To
		}
	}
	return engine.NoSquare
}

// Senses are rule ideas that let a player see more through the fog.
var Senses = []string{"", "attackers", "center", "centerPawn", "centerPeek"}

// Outposts are the two centre squares of c's opponent's half: d5 and e5 for
// White, d4 and e4 for Black.
func Outposts(c engine.Color) [2]engine.Square {
	if c == engine.White {
		return [2]engine.Square{engine.SquareAt(3, 4), engine.SquareAt(4, 4)}
	}
	return [2]engine.Square{engine.SquareAt(3, 3), engine.SquareAt(4, 3)}
}

// OutpostHolder is c's piece that gives it the whole board under sense, or
// NoSquare. arrived is the square c's last move landed on (for centerPeek).
func OutpostHolder(pos *engine.Position, c engine.Color, sense string, arrived engine.Square) engine.Square {
	for _, sq := range Outposts(c) {
		pc := pos.Board[sq]
		if !pc.Is(c) {
			continue
		}
		switch sense {
		case "center":
			return sq
		case "centerPawn":
			if pc.Kind() == engine.Pawn {
				return sq
			}
		case "centerPeek":
			if sq == arrived {
				return sq
			}
		}
	}
	return engine.NoSquare
}

// vision is what color c sees under fog, with sense:
//   - "": what its pieces see, as on the site.
//   - "attackers": also every enemy piece that attacks c's king, wherever it is.
//   - "center": the whole board while any of c's pieces stands on an outpost
//     (the centre of the opponent's half, see Outposts).
//   - "centerPawn": the same, but only a pawn on an outpost counts.
//   - "centerPeek": the whole board only on the turn right after a piece
//     lands on an outpost: one look, then it must leave and come back.
//
// arrived is the square c's last move landed on.
func vision(pos *engine.Position, c engine.Color, sense string, arrived engine.Square) [64]bool {
	seen := pos.Vision(c)
	switch sense {
	case "attackers":
		if king := pos.KingSquare(c); king != engine.NoSquare {
			for sq, pc := range pos.Board {
				if !pc.Is(c.Other()) || seen[sq] {
					continue
				}
				for _, m := range pos.PieceMoves(engine.Square(sq)) {
					if m.To == king {
						seen[sq] = true
						break
					}
				}
			}
		}
	case "center", "centerPawn", "centerPeek":
		if OutpostHolder(pos, c, sense, arrived) != engine.NoSquare {
			for i := range seen {
				seen[i] = true
			}
		}
	}
	return seen
}

// value is a piece's worth in pawns. The king is worth the game.
var value = [...]int{engine.NoKind: 0, engine.Pawn: 1, engine.Knight: 3, engine.Bishop: 3, engine.Rook: 5, engine.Queen: 9, engine.King: 1000}

type randomBot struct{ rng *rand.Rand }

func (b *randomBot) Move(v *View) engine.Move { return v.Legal[b.rng.IntN(len(v.Legal))] }

// greedyBot eats the king when it can see it, otherwise takes the most
// valuable piece it can without losing more, otherwise makes a move that
// no piece it can see attacks.
type greedyBot struct{ rng *rand.Rand }

func (b *greedyBot) Move(v *View) engine.Move {
	pos := beliefPosition(v, nil)
	best, bestScore := engine.Move{}, -1<<30
	for _, m := range v.Legal {
		gain := value[v.Board[m.To].Kind()] * 10
		if v.Board[m.To].Kind() == engine.King {
			return m
		}
		after := pos.Apply(m)
		if after.Attacked(m.To, v.Me.Other()) {
			gain -= value[v.Board[m.From].Kind()] * 10
		}
		if m.Promo == engine.Queen {
			gain += 80
		}
		score := gain*100 + b.rng.IntN(100)
		if score > bestScore {
			best, bestScore = m, score
		}
	}
	return best
}

// beliefPosition is the board as v's player believes it: what they see, plus
// pieces they remember (ghosts), with v.Me to move.
func beliefPosition(v *View, ghosts *[64]engine.Piece) engine.Position {
	p := engine.Position{Board: v.Board, Turn: v.Me, Castling: v.Castling, EP: engine.NoSquare, Fullmove: 1}
	p.KingCapture, p.JumpKing, p.BackPawns, p.Hill = v.Rules.KingCapture, v.Rules.JumpKing, v.Rules.BackPawns, v.Rules.Hill
	p.CenterVision = v.Rules.CenterVision
	if ghosts != nil {
		for i, g := range ghosts {
			if g != engine.NoPiece && !v.Seen[i] {
				p.Board[i] = g
			}
		}
	}
	return p
}
