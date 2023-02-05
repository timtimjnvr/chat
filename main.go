package main

import (
	"chat/conn"
	"chat/crdt"
	"chat/parsestdin"

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
		myNamePtr = flag.String("u", "Tim", "nickname used in chats")
	)
	flag.Parse()

	var (
		id, _   = uuid.NewUUID()
		myInfos = crdt.NewNodeInfos(id, *myNamePtr, *myAddrPtr, *myPortPtr)

		sigc     = make(chan os.Signal, 1)
		shutdown = make(chan struct{})

		wgHandleNodes = sync.WaitGroup{}
		wgListen      = sync.WaitGroup{}
		wgHandleChats = sync.WaitGroup{}
		wgHandleStdin = sync.WaitGroup{}
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
		newNodes  = make(chan net.Conn, conn.MaxSimultaneousConnections)
		toSend    = make(chan []byte, conn.MaxSimultaneousMessages)
		toExecute = make(chan []byte, conn.MaxSimultaneousMessages)
	)

	wgHandleNodes.Add(1)
	go conn.HandleNodes(&wgHandleNodes, newNodes, toSend, toExecute, shutdown)

	wgListen.Add(1)
	go conn.ListenAndServe(&wgListen, *myAddrPtr, *myPortPtr, newNodes, shutdown)

	wgHandleChats.Add(1)
	go crdt.HandleChats(&wgHandleChats, myInfos, toSend, toExecute, shutdown)

	wgHandleStdin.Add(1)
	go parsestdin.HandleStdin(&wgHandleStdin, myInfos, newNodes, toExecute, shutdown)

	for {
		select {
		case <-sigc:
			close(shutdown)
			return
		}
	}
}
