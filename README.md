# chat 
![example workflow](https://github.com/timtimjnvr/chat/actions/workflows/build.yml/badge.svg)
![example workflow](https://github.com/timtimjnvr/chat/actions/workflows/tag-releases.yml/badge.svg)

Decentralized P2P chat built in golang.

## Initial dev -> v1.0.0

- client - client connections (2 nodes) :
  - Open a new chat discussion given an ip address and a port.
  - Basic text messages exchanges.
  - List current discussions.
  - Switch discussion.
  - Close a chat discussion.
  - Order messages with operation based CRDTS.

## run
```
go build && ./chat -p=8080
```

## commands
```
/connnect addr port : open a new chat with a client.
/msg hello, friend ! : send "hello, friend" in the current discussion.
/close : close the current discussion.
```

## Doc
- Node managment : [CRDTs choices](doc/crdt.md)

