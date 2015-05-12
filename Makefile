default: build

clean:
	-rm -r ./build
	gopm clean -v

test:
	gopm test ./...

update:
	gopm update -v

install:
	gopm install -v

format:
	go fmt ./...

simplify:
	gofmt -l -w -s ./**/*.go

vet:
	go vet ./...

lint:
	golint .

build: simplify format vet lint
	#go build -o ./build/grainbot
	gopm build -o ./build/grainbot

open:
	./build/grainbot

run: build open

release: simplify format vet lint test
	GOOS=linux GOARCH=386 go build -o ./build/grainbot_linux

.PHONY: build format run test clean install vet update lint release simplify
