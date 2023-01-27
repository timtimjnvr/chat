package main

import (
	"chat/conn"
	"chat/crdt"
	"chat/linked"
	"chat/node"
	parsestdin "chat/parsestdin"
	"flag"
	"github.com/google/uuid"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

const (
	transportProtocol       = "tcp"
	localhost               = "localhost"
	localhostDecimalPointed = "127.0.0.1"

	maxSimultaneousConnections = 1000
	maxMessagesStdin           = 100

	noDiscussionSelected = "you must be in a discussion to send a message"
)

func main() {
	myPortPtr := flag.String("p", "8080", "port number used to accept conn")
	myAddrPtr := flag.String("a", "", "address used to accept conn")
	myNamePtr := flag.String("u", "Tim", "address used to accept conn")
	flag.Parse()

	var (
		currentChat   = crdt.NewChat(*myNamePtr)
		nodes         = linked.NewList()
		chats         = linked.NewList()
		sigc          = make(chan os.Signal, 1)
		shutdown      = make(chan struct{})
		portAccept    = *myPortPtr
		addressAccept = *myAddrPtr
		wgListen      = sync.WaitGroup{}
		wgReadStdin   = sync.WaitGroup{}

		stdin           = make(chan string, maxMessagesStdin)
		newNodes        = make(chan *node.Node, maxSimultaneousConnections)
		connectionsDone = make(chan uuid.UUID, maxSimultaneousConnections)
	)

	defer func() {
		wgReadStdin.Wait()
		wgListen.Wait()
		// TODO Stop all running nodes
		log.Println("[INFO] program shutdown")
	}()

	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	_ = chats.Add(currentChat)
	wgListen.Add(1)
	go conn.ListenAndServe(&wgListen, newNodes, shutdown, transportProtocol, addressAccept, portAccept)

	wgReadStdin.Add(1)
	go parsestdin.ReadStdin(&wgReadStdin, stdin, shutdown)

	for {
		select {
		case <-sigc:
			close(shutdown)
			return

		/* Save the connection and handle connection*/
		case newNode := <-newNodes:
			nodes.Add(newNode)
			newNode.Business.Wg.Add(1)
			go conn.HandleConnection(newNode, connectionsDone, newNode.Business.Shutdown)

		/* Node done */
		case id := <-connectionsDone:
			nodes.Delete(id)

		/* Command input */
		case line := <-stdin:
			cmd, err := parsestdin.NewCommand(line)
			if err != nil {
				log.Println("[ERROR] ", err)
			}

			args := cmd.GetCommandArgs()

			switch typology := cmd.GetCommandType(); typology {
			case parsestdin.CreateChatCommandType:
				var (
					chatName = args[parsestdin.ChatRoomArg]
					newChat  = crdt.NewChat(chatName)
				)

				chats.Add(newChat)

			case parsestdin.ConnectCommandType:
				var (
					addr     = args[parsestdin.AddrArg]
					chatRoom = args[parsestdin.ChatRoomArg]
					newConn  net.Conn
					pt       int
				)

				pt, err = strconv.Atoi(args[parsestdin.PortArg])
				if err != nil {
					log.Println(err)
				}

				if addr == localhost || addr == localhostDecimalPointed {
					addr = ""
				}

				/* Open connection */
				newConn, err = conn.OpenConnection(transportProtocol, addr, pt)
				if err != nil {
					log.Println("[ERROR] ", err)
					continue
				}

				/* Saves the new connection into newNode List */
				newNode := node.NewNode(newConn)
				newNode.Business.Wg.Add(1)
				go conn.HandleConnection(newNode, connectionsDone, newNode.Business.Shutdown)

				/* Add the newNode to chatRoom */
				if chatRoom == *myNamePtr {
					currentChat.AddNode(newNode)
				}

				/* sync -> add the new newNode to other nodes */

			case parsestdin.MsgCommandType:
				content := args[parsestdin.MessageArg]
				if currentChat == nil {
					log.Println(noDiscussionSelected)
					continue
				}

				/* Add the message to discussion */
				message := crdt.NewMessage(*myNamePtr, content)
				currentChat.AddMessage(message)

				syncMessage := crdt.NewOperation(crdt.AddMessage, message.ToRunes())
				bytesSyncMessage := []byte(string(syncMessage.ToRunes()))
				currentChat.Send(bytesSyncMessage)

			case parsestdin.CloseCommandType:
				/* TODO
				leave discussion and gracefully shutdown connection with all nodes
				*/

			case parsestdin.QuitCommandType:
				/* TODO
				leave all discussions and gracefully shutdown connection with all nodes
				*/
			}
		}
	}
}
