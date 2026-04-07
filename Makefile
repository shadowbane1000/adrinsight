.PHONY: lint test build check clean eval

GO_TAGS = -tags fts5

lint:
	golangci-lint run --build-tags fts5 ./...

test:
	go test -short $(GO_TAGS) ./...

build:
	go build $(GO_TAGS) -o adr-insight ./cmd/adr-insight

check: lint test build
	@echo "All checks passed."

eval: build
	./adr-insight eval

clean:
	rm -f adr-insight adr-insight.db
