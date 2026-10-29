# Eat the King bots

Bots for **Eat the King**, the fog-of-war capture-the-king game from [fow-chess](https://github.com/streamlinedesign/fow-chess), and a simulator that plays them against each other in bulk to learn how the game behaves. Pure Go, no dependencies: no server, browser or database.

- `engine/`: the rules (moves, fog of war vision, king capture, setups' starting positions), copied from `server/engine` in fow-chess.
- `bots/`: the site's four bots, weakest first: `rex` (Rex inutilis, random), `vercingetorix` (greedy), `ramesses` (Ramesses II, a short search) and `solomon` (a deeper search that values vision). Copied from `server/bots` in fow-chess.
- `sim/`: plays games and records them. It adds four simpler bots of its own: `random`; `greedy`, which eats a king it sees and otherwise takes what it can without losing more; `hunter`, which looks two plies ahead on what it sees plus where it last saw enemy pieces; and `wary`, a `hunter` that also keeps lines into its king from the fog covered. A new bot is one entry in `sim.Bots`.
- `cmd/simulate/`: the command.

Every bot gets only what a player gets: its fogged board and its legal moves.

## Running it

Needs Go 1.24 or newer ([go.dev/dl](https://go.dev/dl/)).

```sh
make simulate                                  # plays until Ctrl-C, adding every game to games.jsonl
make simulate ARGS="-games 100000 -fog on"     # a set number of games
make summary                                   # the summary of everything in games.jsonl so far
go run ./cmd/simulate -h                       # every option
```

Without `make` (on Windows, say): `go run ./cmd/simulate -games 0`.

A run cycles through every pairing of `-bots` (each as White and as Black), every `-modes` (`standard`, `backRow`, `battleship`) and fog on and off, so the mix stays balanced however long it runs. It uses every CPU core (`-workers` to change that). Ctrl-C finishes the games in hand and prints the summary; running again adds to the same file.

- `-bots ramesses,solomon` (the default) picks who plays. `solomon` plays about 10 games a second per core; the simulator's own bots play hundreds.
- `-fog on|off|both` (default both).
- `-guarded` makes Battleship setups careful: the king on its back two rows with every line out of it blocked by its own pieces. Otherwise Battleship and Back row setups are random, and random Battleship setups hand out a lot of first-move wins.
- `-mods checkmate,jumpKing,backPawns,hill` plays the Mods page rules.
- `-moves=false` leaves the moves out, for very long runs: a game with moves is about 500 bytes, so ten million games is about 5 GB.

## What a game record holds

Each line of `games.jsonl` is one game: the bots, mode, fog and rules; the result and how it ended; its length in plies; whether it was a Sniper Win (won on the winner's first move); which piece ate the king and whether the loser could see it just before (`killerHidden`); how many times the losing king moved; captures and material left for each side; the first ply with a capture and the first on which either side saw an enemy piece; how much of the board each side saw on average; and the start position and moves, so any game can be replayed. Game `i` of a run uses seed `-seed + i`, and the same seed always plays the same game.

## Keeping in step with fow-chess

`engine/` and `bots/` are copies. When the rules or bots change in fow-chess, copy the folders over again and change the import paths from `chess/` to `github.com/streamlinedesign/chess-bots/`.
