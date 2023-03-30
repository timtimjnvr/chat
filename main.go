package main

import (
	"github/timtimjnvr/chat/conn"
	"github/timtimjnvr/chat/crdt"
	"github/timtimjnvr/chat/parsestdin"

	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	// Program arguments
	var (
		myPortPtr = flag.String("p", "8080", "port number used to accept conn")
		myAddrPtr = flag.String("a", "", "address used to accept conn")
		myNamePtr = flag.String("u", "tim", "nickname used in chats")
	)
	flag.Parse()

	var (
		myInfos = crdt.NewNodeInfos(*myAddrPtr, *myPortPtr, *myNamePtr)

		sigc     = make(chan os.Signal, 1)
		shutdown = make(chan struct{})

		orch        = newOrchestrator(myInfos)
		nodeHandler = conn.NewNodeHandler(shutdown)

		wgListen              = sync.WaitGroup{}
		wgHandleChats         = sync.WaitGroup{}
		wgHandleStdin         = sync.WaitGroup{}
		wgInitNodeConnections = sync.WaitGroup{}
		lock                  = sync.Mutex{}
		isListening           = sync.NewCond(&lock)
	)

	defer func() {
		wgHandleStdin.Wait()
		wgHandleChats.Wait()
		wgListen.Wait()
		nodeHandler.Wg.Wait()
		wgInitNodeConnections.Wait()
		log.Println("[INFO] program shutdown")
	}()

	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	var (
		outGoingCommands = make(chan parsestdin.Command)
		joinChatCommands = make(chan parsestdin.Command)
		newConnections   = make(chan net.Conn)
		toSend           = make(chan *crdt.Operation)
		toExecute        = make(chan *crdt.Operation)
	)

	// listen for new connections
	wgListen.Add(1)
	isListening.L.Lock()
	go conn.Listen(&wgListen, isListening, *myAddrPtr, *myPortPtr, newConnections, shutdown)
	isListening.Wait()

	// create connections with new nodes
	wgInitNodeConnections.Add(1)
	go conn.InitConnections(&wgInitNodeConnections, myInfos, joinChatCommands, newConnections, shutdown)

	// handle new connections until closure
	nodeHandler.Wg.Add(1)
	go nodeHandler.Start(newConnections, toSend, toExecute)
	defer nodeHandler.Wg.Wait()

	// execute and propagates commands & operations to maintain chat data consistency between nodes
	wgHandleChats.Add(1)
	go orch.handleChats(&wgHandleChats, outGoingCommands, toExecute, toSend, shutdown)

	// extract commands from stdin input
	wgHandleStdin.Add(1)
	go parsestdin.HandleStdin(&wgHandleStdin, os.Stdin, *myInfos, outGoingCommands, joinChatCommands, shutdown)

	for {
		select {
		case <-sigc:
			close(shutdown)
			return
		}
	}
}
