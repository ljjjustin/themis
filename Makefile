.PHONY: all

SRCS = $(shell git ls-files '*.go' | grep -v '^vendor/')

default: binary

fmt:
	gofmt -s -l -w $(SRCS)
