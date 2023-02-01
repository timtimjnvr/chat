package main

import (
	"chat/conn"
	"chat/linked"
	"chat/node"
	parsestdin "chat/parsestdin"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
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

	var (
		myPortPtr = flag.String("p", "8080", "port number used to accept conn")
		myAddrPtr = flag.String("a", "", "address used to accept conn")
		myNamePtr = flag.String("u", "Tim", "address used to accept conn")
	)
	flag.Parse()

	var (
		nodes         = linked.NewList()
		chats         = linked.NewList()
		sigc          = make(chan os.Signal, 1)
		shutdown      = make(chan struct{})
		portAccept    = *myPortPtr
		addressAccept = *myAddrPtr
		myInfos       = node.NewNodeInfos(*myNamePtr, addressAccept, portAccept)
		wgOrchestrate = sync.WaitGroup{}
		wgListen      = sync.WaitGroup{}
		wgReadStdin   = sync.WaitGroup{}

		stdin           = make(chan []byte, maxMessagesStdin)
		fromConnections = make(chan []byte, maxSimultaneousConnections)
		newConnections  = make(chan net.Conn, maxSimultaneousConnections)
		// newNodes        = make(chan *node.Node, maxSimultaneousConnections)
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

	wgListen.Add(1)
	go conn.ListenAndServe(&wgListen, newConnections, shutdown, transportProtocol, addressAccept, portAccept)

	wgReadStdin.Add(1)
	go parsestdin.ReadStdin(&wgReadStdin, stdin, shutdown)

	wgOrchestrate.Add(1)
	go orchestrate(&wgOrchestrate, myInfos, stdin, fromConnections, newConnections, chats, nodes, shutdown)

	// go display(chats, refresh <-chan uuid.UUID)

	for {
		select {
		case <-sigc:
			close(shutdown)
			return
		}
	}
}
