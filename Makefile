.PHONY: unit-test
unit-test:
	go test -v ./...

.PHONY: fmt
fmt:
	go fmt ./...
