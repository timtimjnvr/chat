package main

import (
	"github/timtimjnvr/chat/conn"
	"github/timtimjnvr/chat/crdt"
	"github/timtimjnvr/chat/parsestdin"

	"flag"
	"github.com/google/uuid"
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
		id, _   = uuid.NewUUID()
		myInfos = crdt.NewNodeInfos(id, *myAddrPtr, *myPortPtr, *myNamePtr)

		sigc     = make(chan os.Signal, 1)
		shutdown = make(chan struct{})

		wgHandleNodes         = sync.WaitGroup{}
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
		wgHandleNodes.Wait()
		wgInitNodeConnections.Wait()
		log.Println("[INFO] program shutdown")
	}()

	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	var (
		outGoingCommands = make(chan parsestdin.Command, parsestdin.MaxMessagesStdin)
		joinChatCommands = make(chan parsestdin.Command, conn.MaxSimultaneousConnections)
		newConnections   = make(chan net.Conn, conn.MaxSimultaneousConnections)
		toSend           = make(chan crdt.Operation, conn.MaxSimultaneousMessages)
		toExecute        = make(chan crdt.Operation, conn.MaxSimultaneousMessages)
	)

	// listen for new connections
	wgListen.Add(1)
	isListening.L.Lock()
	go conn.Listen(&wgListen, isListening, *myAddrPtr, *myPortPtr, newConnections, shutdown)
	isListening.Wait()

	// create connections with new nodes
	wgInitNodeConnections.Add(1)
	go conn.InitNodeConnections(&wgInitNodeConnections, myInfos, joinChatCommands, newConnections, shutdown)

	// handle new connections from creation to closure
	wgHandleNodes.Add(1)
	go conn.HandleNodes(&wgHandleNodes, newConnections, toSend, toExecute, shutdown)

	// extract commands from stdin input
	wgHandleStdin.Add(1)
	go parsestdin.HandleStdin(&wgHandleStdin, os.Stdin, myInfos, outGoingCommands, shutdown)

	// execute and propagates operations to maintain chat data consistency between nodes
	wgHandleChats.Add(1)
	var orch = NewOrchestrator(myInfos)
	go orch.HandleChats(&wgHandleChats, outGoingCommands, toSend, toExecute, shutdown)

	for {
		select {
		case <-sigc:
			close(shutdown)
			return
		}
	}
}
