package conn

import (
	"fmt"
	"github.com/pkg/errors"
	"github/timtimjnvr/chat/crdt"
	"github/timtimjnvr/chat/parsestdin"
	"log"
	"net"
	"strconv"
	"sync"
)

const (
	localhost               = "localhost"
	localhostDecimalPointed = "127.0.0.1"
	transportProtocol       = "tcp"

	maxMessageSize = 1000
)

func CreateConnections(wg *sync.WaitGroup, isReady *sync.Cond, myInfos *crdt.NodeInfos, joinChatCommands chan parsestdin.Command, newConnections chan net.Conn, shutdown <-chan struct{}) {
	var (
		c                     net.Conn
		wgInitNodeConnections = sync.WaitGroup{}
		wgClosure             = sync.WaitGroup{}
		err                   error
	)

	wgInitNodeConnections.Add(1)
	go Connect(&wgInitNodeConnections, myInfos, joinChatCommands, newConnections, shutdown)

	defer func() {
		close(newConnections)
		isReady.Signal()
		wgClosure.Wait()
		wgInitNodeConnections.Wait()
		wg.Done()
	}()

	ln, err := net.Listen(transportProtocol, fmt.Sprintf("%s:%s", myInfos.Address, myInfos.Port))
	if err != nil {
		log.Fatal("[ERROR]", err)
	}

	wgClosure.Add(1)
	go handleClosure(&wgClosure, ln, shutdown)
	isReady.Signal()

	for {
		// extracts the first connection on the listener queue
		c, err = ln.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			log.Fatal("[ERROR]", err)
		}

		newConnections <- c
	}
}

func handleClosure(wg *sync.WaitGroup, ln net.Listener, shutdown <-chan struct{}) {
	<-shutdown
	err := ln.Close()
	if err != nil {
		log.Fatal(err)
	}

	wg.Done()
}

func Connect(wg *sync.WaitGroup, myInfos *crdt.NodeInfos, newJoinChatCommands <-chan parsestdin.Command, newConnections chan<- net.Conn, shutdown <-chan struct{}) {
	defer func() {
		wg.Done()
	}()

	for {
		select {
		case <-shutdown:
			return
		case joinChatCommand := <-newJoinChatCommands:
			args := joinChatCommand.GetArgs()
			var (
				addr     = args[parsestdin.AddrArg]
				chatRoom = args[parsestdin.ChatRoomArg]
			)

			// check if port is an int
			pt, err := strconv.Atoi(args[parsestdin.PortArg])
			if err != nil {
				log.Println(err)
			}

			/* Open conn */
			var newConn net.Conn
			newConn, err = openConnection(addr, strconv.Itoa(pt))
			if err != nil {
				log.Println("[ERROR] ", err)
				break
			}

			// init joining process
			_, err = newConn.Write(crdt.NewOperation(crdt.JoinChatByName, chatRoom, myInfos).ToBytes())
			if err != nil {
				log.Println("[ERROR] ", err)
			}

			newConnections <- newConn
		}
	}
}

func openConnection(ip string, port string) (net.Conn, error) {
	if ip == localhost || ip == localhostDecimalPointed {
		ip = ""
	}

	conn, err := net.Dial(transportProtocol, fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		return nil, err
	}

	return conn, nil
}
