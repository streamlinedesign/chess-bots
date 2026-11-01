package sim

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/streamlinedesign/chess-bots/engine"
)

// Every bot finishes games in every mode, with and without fog, and a seed
// plays the same game every time.
func TestPlay(t *testing.T) {
	seed := uint64(1)
	for name := range Bots {
		if name == "solomon" && testing.Short() {
			continue // about 100 ms a move
		}
		for _, mode := range []string{"standard", "backRow", "battleship"} {
			for _, fog := range []bool{true, false} {
				for _, rules := range []engine.Rules{{KingCapture: true}, {}, {KingCapture: true, Hill: true, JumpKing: true, BackPawns: true}} {
					seed++
					c := Config{White: name, Black: "greedy", Mode: mode, Fog: fog, Rules: rules, Seed: seed, Moves: true, Guarded: seed%2 == 0}
					a, err := Play(c)
					if err != nil {
						t.Fatal(err)
					}
					b, _ := Play(c)
					if a.Moves != b.Moves || a.Result != b.Result {
						t.Fatalf("%s %s: seed %d played two different games", name, mode, seed)
					}
					if a.Result == "*" {
						t.Fatalf("%s %s: game didn't finish", name, mode)
					}
					replay(t, a, rules)
				}
			}
		}
	}
}

// replay plays a record's moves through the engine and checks it ends the same way.
func replay(t *testing.T, r Record, rules engine.Rules) {
	t.Helper()
	start := r.StartFEN
	if start == "" {
		start = engine.StartFEN
	}
	g, err := engine.NewGameFromFEN(start, rules)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range strings.Fields(r.Moves) {
		if _, err := g.PlayUCI(m); err != nil {
			t.Fatalf("seed %d: %s: %v", r.Seed, m, err)
		}
	}
	if g.Over() && (g.Outcome().String() != r.Result || g.Method().String() != r.Method) {
		t.Fatalf("seed %d: replay ended %s by %s, record says %s by %s", r.Seed, g.Outcome(), g.Method(), r.Result, r.Method)
	}
}

// Random setups put each army only where the site's setup modes allow.
func TestRandomSetup(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	for range 200 {
		fen, err := RandomSetup("battleship", true, true, rng)
		if err != nil {
			t.Fatal(err)
		}
		p, err := engine.ParseFEN(fen)
		if err != nil {
			t.Fatal(err)
		}
		var count [16]int
		for sq, pc := range p.Board {
			if pc.IsEmpty() {
				continue
			}
			count[pc]++
			rank := engine.Square(sq).Rank()
			if pc.Color() == engine.White && rank > 3 || pc.Color() == engine.Black && rank < 4 {
				t.Fatalf("%s: piece outside its rows", fen)
			}
			if pc.Kind() == engine.King && engine.OnHill(engine.Square(sq)) {
				t.Fatalf("%s: king starts on the hill", fen)
			}
		}
		if count[engine.NewPiece(engine.White, engine.Pawn)] != 8 || count[engine.NewPiece(engine.Black, engine.Queen)] != 1 {
			t.Fatalf("%s: wrong army", fen)
		}
	}
	fen, _ := RandomSetup("backRow", false, false, rng)
	if !strings.Contains(fen, "/pppppppp/8/8/8/8/PPPPPPPP/") {
		t.Fatalf("back row setup moved pawns: %s", fen)
	}
}

// A fogged view shows nothing the player's pieces can't see.
func TestViewHidesFog(t *testing.T) {
	g := engine.NewGameWithRules(engine.Rules{KingCapture: true})
	v := viewOf(g, engine.White, true, "", engine.Rules{KingCapture: true})
	for i, pc := range v.Board {
		if pc.Is(engine.Black) {
			t.Fatalf("white sees a black piece on %s at the start", engine.Square(i))
		}
	}
	if len(v.Legal) != 20 {
		t.Fatalf("want 20 legal moves, got %d", len(v.Legal))
	}
}

// With the attackers sense, a hidden piece attacking the king shows; one
// that doesn't attack it stays hidden.
func TestSenseAttackers(t *testing.T) {
	// The black queen on a5 attacks the white king on e1 through the empty
	// diagonal; the black rook on h8 attacks nothing of White's.
	pos, _ := engine.ParseFEN("4k2r/8/8/q7/8/8/8/4K3 w - - 0 1")
	a5, _ := engine.ParseSquare("a5")
	h8, _ := engine.ParseSquare("h8")
	if plain := vision(&pos, engine.White, "", engine.NoSquare); plain[a5] {
		t.Fatal("the white king alone shouldn't see a5")
	}
	seen := vision(&pos, engine.White, "attackers", engine.NoSquare)
	if !seen[a5] || seen[h8] {
		t.Fatalf("attackers sense: a5 seen %v (want true), h8 seen %v (want false)", seen[a5], seen[h8])
	}
}

// The centre senses give the whole board from d5/e5 (White) or d4/e4 (Black).
func TestSenseCenter(t *testing.T) {
	sq := func(s string) engine.Square { q, _ := engine.ParseSquare(s); return q }
	count := func(v [64]bool) int {
		n := 0
		for _, ok := range v {
			if ok {
				n++
			}
		}
		return n
	}
	// A white knight on d5 and a black pawn on e4.
	pos, _ := engine.ParseFEN("4k3/8/8/3N4/4p3/8/8/4K3 w - - 0 1")
	if n := count(vision(&pos, engine.White, "center", engine.NoSquare)); n != 64 {
		t.Errorf("knight on d5, center: White sees %d squares, want 64", n)
	}
	if n := count(vision(&pos, engine.White, "centerPawn", engine.NoSquare)); n == 64 {
		t.Error("centerPawn: a knight shouldn't count")
	}
	if n := count(vision(&pos, engine.Black, "centerPawn", engine.NoSquare)); n != 64 {
		t.Errorf("pawn on e4, centerPawn: Black sees %d squares, want 64", n)
	}
	if n := count(vision(&pos, engine.White, "centerPeek", sq("d5"))); n != 64 {
		t.Errorf("centerPeek right after landing: White sees %d squares, want 64", n)
	}
	if n := count(vision(&pos, engine.White, "centerPeek", sq("e1"))); n == 64 {
		t.Error("centerPeek: a piece that stayed shouldn't give the board again")
	}
	// Black's own half doesn't count for White.
	pos, _ = engine.ParseFEN("4k3/8/8/8/3N4/8/8/4K3 w - - 0 1")
	if n := count(vision(&pos, engine.White, "center", engine.NoSquare)); n == 64 {
		t.Error("a white piece on d4 shouldn't give White the board")
	}
}
