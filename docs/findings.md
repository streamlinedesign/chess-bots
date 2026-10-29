# First findings: 403,240 bot games of Eat the King

Played on 29 October 2026 with `cmd/simulate` on a 4-core cloud machine: 300,000 games between the simulator's own bots (random, greedy, hunter, wary) in all three modes with fog on and off, 100,000 Battleship games with guarded setups, and 3,240 games between the site's bots Vercingetorix, Ramesses II and Solomon (guarded Battleship setups). Every bot sees only what a player would. Score is from White's side (win 1, draw ½); ± is a 95% margin.

**Every number here describes these bots, not people.** The point is what changes between conditions (fog on or off, mode, setup), which is much sturdier than any single number.

## Does fog of war ruin the stats?

No. It changes them, a lot, but they stay meaningful.

**Skill still wins under fog, by less.** The better bot's score against the weaker one, averaged over both colours:

| Matchup | Mode | Fog on | Fog off |
|---|---|---|---|
| wary vs greedy | Standard | 0.66 | 0.99 |
| wary vs greedy | Back row | 0.57 | 0.97 |
| wary vs greedy | Battleship (guarded) | 0.60 | 0.94 |
| Solomon vs Vercingetorix (120 games) | Standard | 0.58 | 0.95 |

Without fog the stronger bot almost never loses. With fog it wins about 6 or 7 games in 10. Each game carries more luck, so you need more games to tell two players apart, but the gap is clearly there with a few thousand games. For ratings, this means fog games should move ratings less per game, or take more games to settle.

**Fog kills draws and shortens games.** Mirror matches of the careful bot (wary vs wary), Standard: 6.6% draws and a median of 39 plies with fog, against 63% draws and 108 plies without. The site bots show the same: 0% draws with fog, 19% without. Without fog, careful bots shuffle into repetition; with fog, someone gets surprised.

**Most kings die to a piece they couldn't see.** About 78% of king captures in Standard and Back row (75% for the site bots) were made by a piece the loser couldn't see on their last move. The "king in danger" ring only ever warns about the other 22-25%.

**Fog changes who does the killing.** Without fog the queen eats about 55% of kings. With fog it's spread out: queen 28%, rook 27%, bishop 20%, knight 17%, pawn 6% (Standard). Long-range pieces arriving from the dark, and knights, which give no line to see coming.

**Fog rewards bots built for it.** `hunter` beats `greedy` 98-99% without fog and only 50% with it. `wary` is `hunter` plus one idea, keep the lines into your king from the fog covered, and that alone takes it back to 66%. Likewise Solomon, which searches deeper than Ramesses II, scores only 0.37-0.43 against it with fog (0.77-0.95 without): searching deeper on a board that treats the fog as empty squares makes it more confidently wrong. Strength in Eat the King is mostly about handling what you can't see.

## First-move advantage

- **Standard and Back row: none to speak of.** White scores 0.48-0.52 in mirror matches of the simulator's bots, with and without fog (Solomon vs Solomon: 0.53-0.55, from only 120 games each).
- **Battleship: White is clearly favoured.** 0.55-0.56 in mirror matches with guarded setups and fog (about 5,500 games each), 0.54-0.59 without. Solomon vs Solomon: 0.56 with fog, 0.67 without (120 games each).

## Battleship and Sniper Wins

With random Battleship setups, **42% of games are Sniper Wins** (won on the winner's first move) and the median game lasts 3 to 10 plies, because random setups leave kings exposed across the middle. With guarded setups (king on the back two rows, every line out of it blocked by its own pieces) that falls to **0.7%**, and the median game lasts 17-30 plies. So Sniper Wins measure how carelessly people set up, and the bigger Battleship question is White's first-move edge above.

## What a normal PC can do overnight

On this 4-core machine the simulator's own bots played about 840 games a second (a mix that includes the very fast random bot; the careful bots alone run a few hundred a second). The site's bots are much slower because Solomon searches four plies: about 4 games a second for a mix with Solomon in it.

On an 8-core desktop over 10 hours, expect roughly:

- tens of millions of games between the simulator's bots. Use `-moves=false` for runs that size: with moves a game is about 500 bytes, so 10 million games is about 5 GB.
- about 250,000-300,000 games with Solomon in the mix, far more with only Vercingetorix and Ramesses.

That's far more than needed for most questions: a few thousand games per condition gives margins of about ±0.02.

## What to look at next

1. **Make a fog-aware Solomon.** The `wary` idea (count open lines from the fog into your own king) is about 15 lines; adding it to the site's bots should make them stronger, and better bots make every other number here more trustworthy.
2. **Battleship balance.** Measure White's edge with the strongest bots and guarded setups, then try fixes: Black sets up after seeing nothing but moving second with an extra tempo, or White not being allowed to capture on move one.
3. **Setup study.** Since Battleship setups decide so much, have bots try many setups and record which arrangements win: where the king should sit, whether pawns in front matter.
4. **Ratings under fog.** With per-game luck this high, check how many games Glicko-2 needs to rank bots of known strength in the right order, and whether fog games should count for less.
5. **The mods.** Run checkmate rules, Jump King, backward pawns and King of the Hill (`-mods`) with fog to see which of them make the game more or less decisive.
