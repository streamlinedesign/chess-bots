# Plays bot games until Ctrl-C, adding each to games.jsonl, then prints a
# summary. Ctrl-C finishes the games in hand first; run it again to add more.
# Options go in ARGS, e.g. make simulate ARGS="-games 10000 -fog on".
# Every option: go run ./cmd/simulate -h
simulate:
	go run ./cmd/simulate -games 0 $(ARGS)

# The summary of everything in games.jsonl so far.
summary:
	go run ./cmd/simulate -summary games.jsonl

test:
	go test ./...

.PHONY: simulate summary test
