all: rss-service

rss-service:
	go build -o rss-service

test: unit-test integration-test

unit-test:
	go test -v -race -timeout=30s ./...

integration-test:
	go test -v -race -timeout=120s --tags=integration ./...

.PHONY: clean
clean:
	go clean
	rm -f rss-service
