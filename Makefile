TAGS := "netgo osusergo"

.PHONY: all clean test

all:
	go build -tags $(TAGS) -o jrnl.out

test:
	go test ./... -v -cover

clean:
	-rm -f *.out
