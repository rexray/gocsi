all: build

ifneq (on,$(GO111MODULE))
export GO111MODULE := on
endif

PROG := mock

.PHONY: deps
deps:
	go mod download && go mod verify

$(PROG): *.go ./provider/*.go ./service/*.go ../*.go
	@go install .
	go build -o "$@" .

build: deps $(PROG)

clean:
	go clean -i -v .
	rm -f $(PROG)

clobber: clean

.PHONY: clean clobber
