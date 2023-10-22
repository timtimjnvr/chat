BINARY_NAME=chat

build:
	go build -o "${BINARY_NAME}" .

init:
	redis-server > /dev/null &

run:
	test -f "${BINARY_NAME}" && ./${BINARY_NAME} || go run .

test:
	go test ./... -race -timeosut 5m