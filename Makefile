.ONESHELL:
.PHONY: all build install

all: build

build:
	go build -o fabricnetgenerator -ldflags "-X main.xVersion=2.2.0" *.go 

install: 
	go install  -ldflags "-X main.xVersion=2.2.0" github.com/suddutt1/fabricnetgenerator
	