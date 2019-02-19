install:
	go install ./...

clean:
	go clean ./...

test:
	go test ./... -cover

build:
	go build ./...
