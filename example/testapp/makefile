#include .env

PROJECTNAME=TestApp

# Go related variables.
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)
GOFILES=$(wildcard *.go)

# Redirect error output to a file, so we can show it in development mode.
STDERR=/tmp/.$(PROJECTNAME)-stderr.txt

# PID file will keep the process id of the server
PID=/tmp/.$(PROJECTNAME).pid

# Log file will keep the log of the server
Log=/tmp/$(PROJECTNAME).log

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

## target: clean and build
target: go-clean go-build

## install: install to path
install:

## start: Start server
start: start-server
#	bash -c "trap 'make stop' EXIT; $(MAKE) compile start-server watch run='make compile start-server'"

## stop: Stop server
stop: stop-server

## Restart: Stop AND Start server
restart: restart-server

start-server: stop-server
	@echo "  >  $(PROJECTNAME) is available at $(ADDR)"
	@sleep 1
	@-$(GOBIN)/$(PROJECTNAME) > $(Log) 2>&1 & echo $$! > $(PID)
	@cat $(PID) | sed "/^/s/^/  \>  PID: /"

stop-server:
	@-touch $(PID)
	@-kill -15 `cat $(PID)` 2> /dev/null || true
	@-rm $(PID)

# watch: Run given command when code changes. e.g; make watch run="echo 'hey'"
#watch:
#	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) yolo -i . -e vendor -e bin -c "$(run)"


restart-server: stop-server start-server

## compile: Compile the binary.
compile:
	@-touch $(STDERR)
	@-rm $(STDERR)
	@-$(MAKE) -s go-compile 2> $(STDERR)
	@cat $(STDERR) | sed -e '1s/.*/\nError:\n/'  | sed 's/make\[.*/ /' | sed "/^/s/^/     /" 1>&2

# exec: Run given command, wrapped with custom GOPATH. e.g; make exec run="go test ./..."
#exec:
#	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) $(run)

## clean: Clean build files. Runs `go clean` internally.
clean: go-clean

go-compile: go-clean go-get go-build

go-build:
	@echo "  >  Building binary..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build -o $(GOBIN)/$(PROJECTNAME) $(GOFILES)

go-generate:
	@echo "  >  Generating dependency files..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go generate $(generate)

go-get:
	@echo "  >  Checking if there is any missing dependencies..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go get $(get)

go-install:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install $(GOFILES)

go-clean:
	@echo "  >  Cleaning build cache"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean

.PHONY: help
all: help
help: makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo