all: build

PROG := csc

$(PROG): *.go ./cmd/*.go ../*.go
	@go install .
	go build -o "$@" .

build: $(PROG)

clean:
	go clean -i -v .
	rm -f $(PROG)

clobber: clean

.PHONY: clean clobber
