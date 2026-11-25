.PHONY: build check e2e fixtures install grade-selftest

build:
	go build -o bin/cutroom ./cmd/cutroom

check:
	@test -z "$$(gofmt -l .)" || (gofmt -l . && echo "gofmt needed" && exit 1)
	go vet ./...
	go test ./...

fixtures:
	test -f evals/work/fixtures/talk-late.json || go run ./evals/fixtures -out evals/work/fixtures

e2e: build fixtures
	PATH="$(CURDIR)/bin:$$PATH" evals/e2e.sh

install:
	go install ./cmd/cutroom

grade-selftest: fixtures
	evals/grade/selftest.sh
