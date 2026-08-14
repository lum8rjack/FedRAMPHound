BINARY := fedramphound

.PHONY: build

build:
	go build -ldflags="-s -w" -trimpath -o $(BINARY) .

