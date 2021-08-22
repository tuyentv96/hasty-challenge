.DEFAULT_GOAL := all

.PHONY: default
# Go Flags
GOFLAGS ?= $(GOFLAGS:)
# We need to export GOBIN to allow it to be set
# for processes spawned from the Makefile
export GOBIN ?= $(PWD)/bin
GO=go

migrate:
	sql-migrate up
test:
	go test -v -count=1 ./...
wire:
	wire ./cmd