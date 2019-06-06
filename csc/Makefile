all: build

ifneq (on,$(GO111MODULE))
export GO111MODULE := on
endif

PROG := csc

.PHONY: deps
deps:
	go mod download && go mod verify


$(PROG): *.go ./cmd/*.go ../*.go
	@go install .
	go build -o "$@" .

build: deps $(PROG)

clean:
	go clean -i -v .
	rm -f $(PROG)

clobber: clean

.PHONY: clean clobber
