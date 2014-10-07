default: build

clean:
	rm -r ./build

test:
	go test ./...

update:
	go get -u -t ./...

install:
	go install

format:
	go fmt ./...

vet:
	go vet ./...

lint:
	golint .

build: format vet lint
	go build -o ./build/grainbot

open:
	./build/grainbot

run: build open

PHONY: build format run test clean install vet update lint
