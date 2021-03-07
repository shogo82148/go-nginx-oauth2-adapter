.PHONY: all test clean

all:

test:
	go test -v -race ./...

clean:
	rm -rf dist
