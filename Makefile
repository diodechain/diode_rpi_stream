TESTS= $(shell go list ./...)
GOPATH= $(shell go env GOPATH)

.PHONY: default
default: lint seccheck test clean

.PHONY: test
test:
	go test -race $(TESTS)

.PHONY: lint
lint:
	go vet $(TESTS)
	GO111MODULE=on go get honnef.co/go/tools/cmd/staticcheck@2020.1.3
	$(GOPATH)/bin/staticcheck -go 1.15 $(TESTS)

.PHONY: seccheck
seccheck:
	go vet ./...
	GO111MODULE=on go get github.com/securego/gosec/v2/cmd/gosec
	$(GOPATH)/bin/gosec $(TESTS)

.PHONY: clean
clean:
	go clean -cache ./...
	go mod tidy
