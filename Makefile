build:
	go build -trimpath -o bin/replaynet ./cmd/replaynet

test:
	go test ./...

deps-proof:
	go list -m all > deps-proof.txt

repro:
	./scripts/reproducible-build.sh

clean:
	rm -rf bin
