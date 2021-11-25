GOCMD=go
GOBUILD=$(GOCMD) build

.PHONY: build clean

build:
	$(GOBUILD) -o bin/main main.go

clean:
	$(GOCMD) clean
	rm -rf ./bin
