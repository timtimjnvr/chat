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
/chat <room> :                    create a new room named room and enter it.
/join <addr> <port> <chat_room>:  join the room named room (<addr> and <port> identifies a user already in the room).
/msg hello, friend ! :            send "hello, friend" in the current room.
/close :                          exit the current room.
/list :                           display user(s) in the room.
/list_chats :                     display enterred rooms.
/quit :                           kills the program
```

## Dev & debug
- run `go run . -p=<port_number> -u=<your_nickname>` (default port_number 8080)
- test : `go test ./... -race -timeout 30s -coverprofile cover.out`
- open coverage in browser: `go tool cover -html=cover.out`
- see tcp traffic on a port (debugging) : `sudo tcpdump -i lo0 port <port>`

## Doc
- Node management : [CRDTs choices](doc/crdt.md)
- Architecture
![alt text](https://github.com/timtimjnvr/chat/blob/main/doc/architecture.png?raw=true)