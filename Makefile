.PHONY: all test clean

all:

test:
	go test -v -race $(shell glide novendor)

clean:
	-rm -rf vendor
