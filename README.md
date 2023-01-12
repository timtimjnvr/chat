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
go build ./...
./chat -p=8080
```

## commands

```
/connnect addr port : open a new chat with a client.
/msg hello, friend ! : send "hello, friend" in the current discussion.
/close : close the current discussion.
```


