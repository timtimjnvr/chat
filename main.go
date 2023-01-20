package main

import (
	"chat/crdt"
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
	maxMessageSize             = 10000
	maxMessagesStdin           = 100

	noDiscussionSelected = "you must be in a discussion to send a message"
)

func main() {
	myPortPtr := flag.String("p", "8080", "port number used to accept connection")
	myAddrPtr := flag.String("a", "", "address used to accept connection")
	myNamePtr := flag.String("u", "Tim", "address used to accept connection")
	flag.Parse()

	var (
		currentDiscussion = crdt.NewChat(*myNamePtr)
		nodes             = node.NewNodeList()

		sigc          = make(chan os.Signal, 1)
		shutdown      = make(chan struct{})
		portAccept    = *myPortPtr
		addressAccept = *myAddrPtr
		wgListen      = sync.WaitGroup{}
		wgReadStdin   = sync.WaitGroup{}

		stdin           = make(chan string, maxMessagesStdin)
		newConnections  = make(chan net.Conn, maxSimultaneousConnections)
		connectionsDone = make(chan uuid.UUID, maxSimultaneousConnections)
	)

	defer func() {
		wgReadStdin.Wait()
		wgListen.Wait()
		nodes.CloseAndWaitNode()
		log.Println("[INFO] program shutdown")
	}()

	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	wgListen.Add(1)
	go listenAndServe(&wgListen, newConnections, shutdown, transportProtocol, addressAccept, portAccept)

	wgReadStdin.Add(1)
	go readStdin(&wgReadStdin, stdin, shutdown)

	for {
		select {
		case <-sigc:
			close(shutdown)
			return

		/* Save the connection and handle connection*/
		case conn := <-newConnections:
			newNode := node.NewNode(conn)
			nodes.AddNode(newNode)

			newNode.Business.Wg.Add(1)
			go handleConnection(newNode.Business.Wg, newNode.Business.Conn, newNode.Infos.Id, connectionsDone, newNode.Business.Shutdown)

		/* Node done */
		case id := <-connectionsDone:
			nodes.RemoveNode(id)

		/* Command input */
		case line := <-stdin:
			cmd, err := parsestdin.NewCommand(line)
			if err != nil {
				log.Println("[ERROR] ", err)
			}

			args := cmd.GetCommandArgs()

			switch typology := cmd.GetCommandType(); typology {
			case parsestdin.ConnectCommandType:
				var (
					addr = args[parsestdin.AddrArg]
					// chatRoom = args[parsestdin.ChatRoomArg]
					conn net.Conn
					pt   int
				)

				pt, err = strconv.Atoi(args[parsestdin.PortArg])
				if err != nil {
					log.Println(err)
				}

				if addr == localhost || addr == localhostDecimalPointed {
					addr = ""
				}

				// Open connection
				conn, err = openConnection(transportProtocol, addr, pt)
				if err != nil {
					log.Println("[ERROR] ", err)
					continue
				}

				// Saves new connection
				newConnections <- conn

				// ask to enter chat room
				// TODO

			case parsestdin.MsgCommandType:
				content := args[parsestdin.MessageArg]
				if currentDiscussion == nil {
					log.Println(noDiscussionSelected)
					continue
				}

				// Add the message to discussion
				message := crdt.NewMessage(*myNamePtr, content)
				currentDiscussion.AddMessage(message)

				// build operation
				// operation := crdt.GetOperationRunes(crdt.AddMessage, message)

				// sync nodes by sending operation
				// TODO

			case parsestdin.CloseCommandType:
				// TODO : leave discussion and gracefully shutdown connection with all nodes

			case parsestdin.QuitCommandType:
				// TODO : leave all discussions and gracefully shutdown connection with all nodes
			}
		}
	}
}
