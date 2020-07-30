GOCMD=go
GOBUILD=GOOS=linux $(GOCMD) build -ldflags="-d -s -w" -tags netgo -installsuffix netgo

.PHONY: build clean deploy

build:
	$(GOBUILD) -o bin/setwebhook setwebhook/main.go
	$(GOBUILD) -o bin/webhook webhook/main.go

clean:
	$(GOCMD) clean
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
