package main

import (
	"fmt"
	"github/timtimjnvr/chat/conn"
	"github/timtimjnvr/chat/crdt"
	"github/timtimjnvr/chat/orchestrator"
	"github/timtimjnvr/chat/storage"
	"net"
	"os"
	"sync"
)

func start(addr string, port string, name string, stdin *os.File, sigc chan os.Signal, debugModePtr bool) {
	var (
		myInfos            = crdt.NewNodeInfos(addr, port, name)
		shutDown           = make(chan struct{})
		connectionRequests = make(chan conn.ConnectionRequest)
		newConnections     = make(chan net.Conn)
		toSend             = make(chan *crdt.Operation)
		// 2 senders : node handler & orchestrator (operations from stdin)
		toExecute = make(chan *crdt.Operation, 2)

		wgListen      = sync.WaitGroup{}
		wgHandleChats = sync.WaitGroup{}
		lock          = sync.Mutex{}
		isReady       = sync.NewCond(&lock)
		storage       = storage.NewStorage()
		orch          = orchestrator.NewOrchestrator(storage, myInfos)
		nodeHandler   = conn.NewNodeHandler(storage)
	)

	// create connections : tcp connect & listen for incoming connections
	wgListen.Add(1)
	isReady.L.Lock()
	go conn.CreateConnections(&wgListen, isReady, myInfos, connectionRequests, newConnections, shutDown)
	isReady.Wait()

	// handle created connections until closure
	nodeHandler.Wg.Add(1)
	go nodeHandler.Start(newConnections, toSend, toExecute)
	defer nodeHandler.Wg.Wait()

	// maintain chat infos by executing and propagating operations
	wgHandleChats.Add(1)
	go orch.HandleChats(&wgHandleChats, toExecute, toSend)

	// create operations from stdin input
	orch.HandleStdin(stdin, toExecute, connectionRequests, shutDown, sigc)

	wgHandleChats.Wait()
	wgListen.Wait()
	nodeHandler.Wg.Wait()
	fmt.Println("[INFO] program shutdown")

}
