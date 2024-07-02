# chat
![example workflow](https://github.com/timtimjnvr/chat/actions/workflows/build.yml/badge.svg)
![example workflow](https://github.com/timtimjnvr/chat/actions/workflows/tag-releases.yml/badge.svg)

Decentralized P2P in terminal chat built in golang.

## Commands

```
/chat <room> :                    create a new room named room and enter it.
/join <addr> <port> <chat_room> : join the room named room (<addr> and <port> identifies a user already in the room).
/msg <content> :                  send "content" in the current room.
/close :                          exit the current room.
/list :                           display user(s) in the room.
/list_chats :                     display enterred rooms.
/list_users :                     display all connected users.
/switch <chat_room>:              change the current room to <chat_room> (need to be joined).
/quit :                           kills the program
```

## Doc
- Architecture
  ![alt text](https://github.com/timtimjnvr/chat/blob/main/doc/architecture.png?raw=true)
- data exchanges : [CRDTs choices](doc/crdt.md)
- protocol to join & leave :
![alt text](https://github.com/timtimjnvr/chat/blob/main/doc/sequence.png?raw=true)
