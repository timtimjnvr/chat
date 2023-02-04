# chat 
![example workflow](https://github.com/timtimjnvr/chat/actions/workflows/build.yml/badge.svg)
![example workflow](https://github.com/timtimjnvr/chat/actions/workflows/tag-releases.yml/badge.svg)

Decentralized P2P chat built in golang.

## Initial dev -> v1.0.0

- client - client connections (2 nodes) :
  - Open a new discussion given an ip address and a port and the discussion name.
  - Basic text message exchanges.
  - List current users in discussion.
  - Close a chat discussion.
  - Order messages and handle new users with operation based CRDTs.

## Run

```
go build && ./chat -p=<port_number> -u=<your_nickname>
```
Default port_number : 8080

## Commands

```
/chat <chat_room> : create a new chat room named chat_room and enter it.
/connnect <addr> <port> <chat_room>: join the chat room named chat_room.
/msg hello, friend ! : send "hello, friend" in the current chat room.
/close : exit the current chat room.
/list : display user(s) in the room.
/quit kills the program
```

## Doc
- Node managment : [CRDTs choices](doc/crdt.md)

## Todo
- add proper functions for operations handling
- integrate operation building in stdin parser
- handle node closure
