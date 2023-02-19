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

		wgHandleNodes = sync.WaitGroup{}
		wgListen      = sync.WaitGroup{}
		wgHandleChats = sync.WaitGroup{}
		wgHandleStdin = sync.WaitGroup{}
		lock          = sync.Mutex{}
		isListening   = sync.NewCond(&lock)
	)

	defer func() {
		wgHandleStdin.Wait()
		wgHandleChats.Wait()
		wgListen.Wait()
		wgHandleNodes.Wait()
		log.Println("[INFO] program shutdown")
	}()

	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	var (
		newConnections = make(chan net.Conn, conn.MaxSimultaneousConnections)
		toSend         = make(chan crdt.Operation, conn.MaxSimultaneousMessages)
		toExecute      = make(chan crdt.Operation, conn.MaxSimultaneousMessages)
	)

	wgHandleNodes.Add(1)
	go conn.HandleNodes(&wgHandleNodes, newConnections, toSend, toExecute, shutdown)

	wgListen.Add(1)
	isListening.L.Lock()
	go conn.ListenAndServe(&wgListen, isListening, *myAddrPtr, *myPortPtr, newConnections, shutdown)
	isListening.Wait()

	wgHandleChats.Add(1)
	var orchestrator = crdt.NewOrchestrator(myInfos)
	go orchestrator.HandleChats(&wgHandleChats, toSend, toExecute, shutdown)

	wgHandleStdin.Add(1)
	go parsestdin.HandleStdin(&wgHandleStdin, os.Stdin, myInfos, newConnections, toExecute, shutdown)

	for {
		select {
		case <-sigc:
			close(shutdown)
			return
		}
	}
}
