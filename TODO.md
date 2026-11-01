# Todo

## Now

- Run make configure-repo once
- Review PR #2 (pawn vision, centre vision, hidden-attack warning); it changes `engine/`, so fow-chess needs the same change
- Delete the two stale `claude/project-thread-*` branches once their PRs are closed

## Next

- `engine/` (and `bots/`) are copies of fow-chess's `server/engine` and `server/bots`; keep them in step until the engine becomes a Go module, then import it instead
- Add a check that `engine/` matches fow-chess (a script or a test), so drift is caught before a run
- Tests for `cmd/simulate` (flag parsing, summary output)

## Later

- More rule ideas in `-sense` and `-mods` as fow-chess gains them
- Longer runs for the open questions in `docs/findings.md`

## Done

- 0.1.0: engine and bots copies, simulator, findings
