package main

import (
	"github/timtimjnvr/chat/conn"
	"github/timtimjnvr/chat/crdt"
	"github/timtimjnvr/chat/parsestdin"
	"net"

	"flag"
	"log"
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
		myInfos     = crdt.NewNodeInfos(*myAddrPtr, *myPortPtr, *myNamePtr)
		currentChat *crdt.Chat

		sigc             = make(chan os.Signal, 1)
		shutdown         = make(chan struct{})
		joinChatCommands = make(chan parsestdin.Command)
		newConnections   = make(chan net.Conn)
		toSend           = make(chan *crdt.Operation)
		toExecute        = make(chan *crdt.Operation)

		wgListen      = sync.WaitGroup{}
		wgHandleChats = sync.WaitGroup{}
		wgHandleStdin = sync.WaitGroup{}
		lock          = sync.Mutex{}
		isReady       = sync.NewCond(&lock)

		orch        = newOrchestrator(myInfos)
		nodeHandler = conn.NewNodeHandler(shutdown)
	)

	defer func() {
		wgHandleStdin.Wait()
		wgHandleChats.Wait()
		wgListen.Wait()
		nodeHandler.Wg.Wait()
		log.Println("[INFO] program shutdown")
	}()

	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	// create connections : tcp connect & listen for incoming connections
	wgListen.Add(1)
	isReady.L.Lock()
	go conn.CreateConnections(&wgListen, isReady, myInfos, joinChatCommands, newConnections, shutdown)
	isReady.Wait()

	// handle created connections until closure
	nodeHandler.Wg.Add(1)
	go nodeHandler.Start(newConnections, toSend, toExecute)
	defer nodeHandler.Wg.Wait()

	// maintain chat infos by executing and propagating operations
	wgHandleChats.Add(1)
	go orch.handleChats(&wgHandleChats, toExecute, toSend, shutdown)

	// create operations from stdin input
	wgHandleStdin.Add(1)
	go parsestdin.HandleStdin(&wgHandleStdin, myInfos, currentChat, toExecute, joinChatCommands, shutdown)

	for {
		select {
		case <-sigc:
			close(shutdown)
			return
		}
	}
}
