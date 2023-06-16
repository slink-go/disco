.DEFAULT_GOAL := build

help:
	# Commands:
	# make build  - Build the shared object (.so) library files
	# make inmem   - Run the duck example
	# make redis   - Run the from example

build:
	@go build -ldflags "-s -w" -buildmode plugin -o ../../build/inmem.so inmem/registry.go

#inmem: build
#	@go run main/main.go inmem

