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

simplify:
	gofmt -l -w -s ./**/*.go

vet:
	go vet ./...

lint:
	golint .

build: clean simplify format vet lint
	go build -o ./build/grainbot

open:
	./build/grainbot

run: build open

release: clean simplify format vet lint test
	GOOS=linux GOARCH=386 go build -o ./build/grainbot_linux

.PHONY: build format run test clean install vet update lint release simplify
