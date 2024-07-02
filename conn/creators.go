package conn

import (
	"errors"
	"fmt"
	"github/timtimjnvr/chat/crdt"
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

type ConnectionRequest struct {
	targetedPort    string
	targetedAddress string
	chatRoom        string
}

func NewConnectionRequest(port, address, chatRoom string) ConnectionRequest {
	return ConnectionRequest{
		targetedPort:    port,
		targetedAddress: address,
		chatRoom:        chatRoom,
	}
}

func CreateConnections(wg *sync.WaitGroup, isReady *sync.Cond, myInfos *crdt.NodeInfos, incomingConnectionRequests chan ConnectionRequest, newConnections chan net.Conn, shutdown <-chan struct{}) {
	var (
		c                     net.Conn
		wgInitNodeConnections = sync.WaitGroup{}
		wgClosure             = sync.WaitGroup{}
		err                   error
	)

	wgInitNodeConnections.Add(1)
	go InitJoinChatProcess(&wgInitNodeConnections, myInfos, incomingConnectionRequests, newConnections, shutdown)

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic:", r)
		}

		wgInitNodeConnections.Wait()
		isReady.Signal()
		wgClosure.Wait()
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
		} else if err != nil {
			fmt.Println("[ERROR] ", err.Error())
			return
		}

		newConnections <- c
	}
}

func InitJoinChatProcess(wg *sync.WaitGroup, myInfos *crdt.NodeInfos, incomingConnectionRequest <-chan ConnectionRequest, newConnections chan<- net.Conn, shutdown <-chan struct{}) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("[ERROR] ", r)
		}
		wg.Done()
	}()

	for {
		select {
		case <-shutdown:
			return

		case connectionRequest := <-incomingConnectionRequest:
			var (
				addr     = connectionRequest.targetedAddress
				chatRoom = connectionRequest.chatRoom
			)

			// check if targetedPort is an int
			_, err := strconv.Atoi(connectionRequest.targetedPort)
			if err != nil {
				fmt.Println(err)
			}

			/* Open conn */
			var c net.Conn
			c, err = openConnection(addr, connectionRequest.targetedPort)
			if err != nil {
				fmt.Println("[ERROR] ", err)
				break
			}

			// init joining process
			_, err = c.Write(crdt.NewOperation(crdt.JoinChatByName, chatRoom, myInfos).ToBytes())
			if err != nil {
				fmt.Println("[ERROR] ", err)
			}

			newConnections <- c
		}
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

func openConnection(ip string, port string) (net.Conn, error) {
	if ip == localhost || ip == localhostDecimalPointed || ip == "" {
		ip = ""
	}

	conn, err := net.Dial(transportProtocol, fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		return nil, err
	}

	return conn, nil
}
