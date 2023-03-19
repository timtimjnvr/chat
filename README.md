# chat 
![example workflow](https://github.com/timtimjnvr/chat/actions/workflows/build.yml/badge.svg)
![example workflow](https://github.com/timtimjnvr/chat/actions/workflows/tag-releases.yml/badge.svg)

Decentralized P2P chat built in golang.

## Initial dev -> v1.0.0

- client - client connections (2 nodes) :
  - open a new discussion given an ip address and a port and the discussion name.
  - basic text message exchanges.
  - list current users in discussion.
  - close a chat discussion.
  - order messages and handle new users with operation based CRDTs.

## Commands

```
/chat <chat_room> : create a new chat room named chat_room and enter it.
/join <addr> <port> <chat_room>: join the chat room named chat_room.
/msg hello, friend ! : send "hello, friend" in the current chat room.
/close : exit the current chat room.
/list : display user(s) in the room.
/quit kills the program
```

## Development
- run `go run . -p=<port_number> -u=<your_nickname>` (default port_number 8080)
- test : `go test ./... -race -timeout 30s -coverprofile cover.out`
- open coverage in browser: `go tool cover -html=cover.out`

## Doc
- Node managment : [CRDTs choices](doc/crdt.md)
- Architecture
![alt text](https://github.com/timtimjnvr/chat/blob/main/doc/architecture.png?raw=true)

## TODO
- test node handler under stress : many newNodes / node done / messages to send