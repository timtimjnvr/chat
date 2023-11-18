# chat 
![example workflow](https://github.com/timtimjnvr/chat/actions/workflows/build.yml/badge.svg)
![example workflow](https://github.com/timtimjnvr/chat/actions/workflows/tag-releases.yml/badge.svg)

Decentralized P2P chat built in golang.

## Commands

```
/chat <room> :                    create a new room named room and enter it.
/join <addr> <port> <chat_room>:  join the room named room (<addr> and <port> identifies a user already in the room).
/msg hello, friend ! :            send "hello, friend" in the current room.
/close :                          exit the current room.
/list :                           display user(s) in the room.
/list_chats :                     display enterred rooms.
/list_users :                     display all connected users.
/quit :                           kills the program
```

## Next -> v2.0.0
- Multi-user rooms feature (for now only two users in a room)

## Doc
- Node management : [CRDTs choices](doc/crdt.md)
- Architecture
![alt text](https://github.com/timtimjnvr/chat/blob/main/doc/architecture.png?raw=true)
- Joining a chat
![alt text](https://github.com/timtimjnvr/chat/blob/main/doc/joining_sequence.png?raw=true)