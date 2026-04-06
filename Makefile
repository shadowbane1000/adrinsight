.PHONY: lint test build check clean

lint:
	golangci-lint run ./...

test:
	go test -short ./...

build:
	go build -o adr-insight ./cmd/adr-insight

check: lint test build
	@echo "All checks passed."

clean:
	rm -f adr-insight adr-insight.db
