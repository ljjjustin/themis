.PHONY: all

SRCS = $(shell git ls-files '*.go' | grep -v '^vendor/')

default:
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o themis themis.go
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o themisctl themisctl.go

fmt:
	gofmt -s -l -w $(SRCS)
