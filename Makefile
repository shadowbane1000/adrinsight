.PHONY: lint test build check clean eval

lint:
	golangci-lint run ./...

test:
	go test -short ./...

build:
	go build -o adr-insight ./cmd/adr-insight

check: lint test build
	@echo "All checks passed."

eval: build
	./adr-insight eval

clean:
	rm -f adr-insight adr-insight.db
