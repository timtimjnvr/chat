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

## run

```
go build && ./chat -p=<port_number> -u=<your_nickname>
```
Default port_number : 8080

## commands

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

## TODO
- add node to chat room (how to join a chat room by id or by name does the owner has to send chat id after connect)
- clean sync operation building and decoding
