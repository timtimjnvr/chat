package main

import (
	"fmt"
	"github/timtimjnvr/chat/conn"
	"github/timtimjnvr/chat/crdt"
	"github/timtimjnvr/chat/orchestrator"
	"net"
	"os"
	"sync"
)

func start(addr string, port string, name string, stdin *os.File, sigc chan os.Signal, debugModePtr bool) {
	var (
		myInfos            = crdt.NewNodeInfos(addr, port, name)
		shutdown           = make(chan struct{})
		connectionRequests = make(chan conn.ConnectionRequest)
		newConnections     = make(chan net.Conn)
		toSend             = make(chan *crdt.Operation)
		toExecute          = make(chan *crdt.Operation)

		wgListen      = sync.WaitGroup{}
		wgHandleChats = sync.WaitGroup{}
		wgHandleStdin = sync.WaitGroup{}
		lock          = sync.Mutex{}
		isReady       = sync.NewCond(&lock)

		orch        = orchestrator.NewOrchestrator(myInfos)
		nodeHandler = conn.NewNodeHandler(shutdown)
	)

	if debugModePtr {
		orch.SetDebugMode()
		nodeHandler.SetDebugMode()
	}

	defer func() {
		wgHandleStdin.Wait()
		wgHandleChats.Wait()
		wgListen.Wait()
		nodeHandler.Wg.Wait()
		fmt.Println("[INFO] program shutdown")
	}()

	// create connections : tcp connect & listen for incoming connections
	wgListen.Add(1)
	isReady.L.Lock()
	go conn.CreateConnections(&wgListen, isReady, myInfos, connectionRequests, newConnections, shutdown)
	isReady.Wait()

	// handle created connections until closure
	nodeHandler.Wg.Add(1)
	go nodeHandler.Start(newConnections, toSend, toExecute)
	defer nodeHandler.Wg.Wait()

	// maintain chat infos by executing and propagating operations
	wgHandleChats.Add(1)
	go orch.HandleChats(&wgHandleChats, toExecute, toSend, shutdown)

	// create operations from stdin input
	wgHandleStdin.Add(1)
	go orch.HandleStdin(&wgHandleStdin, stdin, toExecute, connectionRequests, shutdown)

	select {
	case <-sigc:
		close(shutdown)
	}
}
