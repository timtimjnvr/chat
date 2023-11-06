build:
	go build  .

help:
	go run . -help

test:
	go test ./... -v -race -timeout 30s

coverage:
	go test ./... -race -timeout 30s -coverprofile cover.out
	go tool cover -html=cover.out

run:
	go run .