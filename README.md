# chat 
P2P chat program built in golang.

## Initial dev : v1.0.0

- client - client connections :
  - Open new chat discussion.
  - Close a chat discussion.
  - Basic text messages exchange.
  - Order messages with CRDTS.

## run

```
cd client
go build client
./client -p=8080
```
```
cd client
./client -p=8081
```

## commands

```
/connnect localhost 8080 : open a new chat with a client listening on localhost:8080.
/msg hello, friend ! : send "hello, friend" in the current discussion.
/close : close the current discussion.
/list : list the current connections.
```


