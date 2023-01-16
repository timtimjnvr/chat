package main

import (
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
	messageMaxSize             = 10000
	maxMessagesStdin           = 100
	maxMessagesConn            = 100

	noDiscussionSelected = "you must be in a discussion to send a message"
)

func main() {
	portPtr := flag.String("p", "8080", "port number used to accept connection")
	addrPtr := flag.String("a", "", "address used to accept connection")
	flag.Parse()

	var (
		currentDiscussion *node.Node
		nodes             = node.NewNodeList()

		sigc          = make(chan os.Signal, 1)
		shutdown      = make(chan struct{})
		portAccept    = *portPtr
		addressAccept = *addrPtr
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

		case conn := <-newConnections:
			currentDiscussion = node.NewChat(conn)
			nodes.AddNode(currentDiscussion)
			currentDiscussion.Infos.Wg.Add(1)
			go handleConnection(currentDiscussion.Infos.Wg, currentDiscussion.Infos.Conn, currentDiscussion.Id, connectionsDone, currentDiscussion.Infos.Shutdown)

		case id := <-connectionsDone:
			nodes.RemoveNode(id)
			currentDiscussion = nil

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

				conn, err = openConnection(transportProtocol, addr, pt)
				if err != nil {
					log.Println("[ERROR] ", err)
					continue
				}

				newConnections <- conn

			case parsestdin.MsgCommandType:
				content := args[parsestdin.MessageArg]
				if currentDiscussion == nil {
					log.Println(noDiscussionSelected)
					continue
				}

				err = sendMessage(currentDiscussion, content)
				if err != nil {
					log.Println("[ERROR] ", err)
				}

			case parsestdin.CloseCommandType:
				currentDiscussion.Stop()

			case parsestdin.ListDiscussionCommandType:
				nodes.Display()

			case parsestdin.SwitchDiscussionCommandType:
				// chatId, _ := strconv.Atoi(args[parsestdin.IdChatArg])
			}
		}
	}
}
