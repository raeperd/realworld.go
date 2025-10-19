default: build lint test 

download:
	go mod download

build: download
	go build ./...

test:
	go test -shuffle=on -race -coverprofile=coverage.txt ./...

lint: download
	golangci-lint run
