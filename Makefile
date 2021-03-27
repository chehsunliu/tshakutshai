COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

.PHONY: test
test:
	go test -v ./...

.PHONY: coverage
coverage:
	go test -v -coverprofile=$(COVERAGE_FILE) ./...
	go tool cover -func $(COVERAGE_FILE)
	go tool cover -html $(COVERAGE_FILE) -o $(COVERAGE_HTML)

.PHONY: fmt
fmt:
	go fmt ./...
