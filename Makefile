GO ?= go
PACKAGE := ./...
GOCACHE ?= $(CURDIR)/.cache/go-build
GOFILES := $(shell find . -type f -name '*.go' -not -path './.git/*' -not -path './.cache/*')

.PHONY: help fmt test vet check clean

help:
	@printf "Available targets:\n"
	@printf "  make fmt    - format Go source files\n"
	@printf "  make test   - run Go tests\n"
	@printf "  make vet    - run go vet\n"
	@printf "  make check  - run fmt, vet, and test\n"
	@printf "  make clean  - remove local build cache\n"

fmt:
	@mkdir -p "$(GOCACHE)"
	$(GO) fmt $(PACKAGE)

test:
	@mkdir -p "$(GOCACHE)"
	GOCACHE="$(GOCACHE)" $(GO) test $(PACKAGE)

vet:
	@mkdir -p "$(GOCACHE)"
	GOCACHE="$(GOCACHE)" $(GO) vet $(PACKAGE)

check: fmt vet test

clean:
	rm -rf .cache coverage.out coverage.html dist bin
