PKGS := $(shell go list ./... | grep -v /vendor)
SRC := $(shell find . -name '*.go')

.PHONY: test
test: format
	go test $(PKGS)

format: $(SRC)
	go fmt -x $(PKGS)
