package main

import (
	"chat/data"
	"flag"
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
)

func main() {
	portPtr := flag.String("p", "8080", "port number used to accept connection")
	addrPtr := flag.String("a", "", "address used to accept connection")
	flag.Parse()

	var (
		currentChat *data.Chat
		chatList    = data.NewChatList()

		sigc          = make(chan os.Signal, 1)
		shutdown      = make(chan struct{})
		portAccept    = *portPtr
		addressAccept = *addrPtr
		wgListen      = sync.WaitGroup{}
		wgReadStdin   = sync.WaitGroup{}

		stdin          = make(chan string, maxMessagesStdin)
		newConnections = make(chan net.Conn, maxSimultaneousConnections)
	)

	defer func() {
		wgReadStdin.Wait()
		wgListen.Wait()
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
			// add to chat list
			currentChat = data.NewChat(conn)
			chatList.AddChat(currentChat)

			go handleConnection(currentChat.Infos.Wg, currentChat.Infos.Conn, currentChat.Infos.Shutdown)
			log.Println(chatList)

		case line := <-stdin:
			cmd, err := parseCommand(line)
			if err != nil {
				log.Println("[ERROR] ", err)
			}

			switch cmd.typology {
			case connectCommandType:
				var (
					addr = cmd.args[addrArg]
					conn net.Conn
					pt   int
				)
				pt, err = strconv.Atoi(cmd.args[portArg])
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

			case msgCommandType:
				content := cmd.args[messageArg]
				err = sendMessage(currentChat, content)
				if err != nil {
					log.Println("[ERROR] ", err)
				}

			case closeCommandType:
				close(currentChat.Infos.Shutdown)

			case listDiscussionCommandType:
				chatList.Display()

			case switchDiscussionCommandType:
				chatId, _ := strconv.Atoi(cmd.args[idChatArg])
				currentChat = chatList.GetChat(chatId)
			}
		}
	}
}
