package conn

import (
	"fmt"
	"github.com/pkg/errors"
	"github/timtimjnvr/chat/crdt"
	"github/timtimjnvr/chat/parsestdin"
	"github/timtimjnvr/chat/reader"
	"log"
	"net"
	"strconv"
	"sync"
)

type (
	node struct {
		slot     uint8
		conn     net.Conn
		Input    chan []byte
		Output   chan []byte
		Wg       *sync.WaitGroup
		Shutdown chan struct{}
	}
)

const (
	localhost               = "localhost"
	localhostDecimalPointed = "127.0.0.1"
	transportProtocol       = "tcp"

	MaxSimultaneousConnections = 100
	MaxSimultaneousMessages    = 100
	MaxMessageSize             = 1000
)

func newNode(conn net.Conn, slot uint8, output chan []byte) node {
	return node{
		slot:     slot,
		conn:     conn,
		Input:    make(chan []byte, MaxSimultaneousMessages),
		Output:   output,
		Wg:       &sync.WaitGroup{},
		Shutdown: make(chan struct{}, 0),
	}
}

func Listen(wg *sync.WaitGroup, isReady *sync.Cond, addr, port string, newConnections chan net.Conn, shutdown chan struct{}) {
	var (
		conn      net.Conn
		wgClosure = sync.WaitGroup{}
		err       error
	)

	defer func() {
		isReady.Signal()
		wgClosure.Wait()
		wg.Done()
	}()

	ln, err := net.Listen(transportProtocol, fmt.Sprintf("%s:%s", addr, port))
	if err != nil {
		log.Fatal("[ERROR]", err)
	}

	wgClosure.Add(1)
	go handleClosure(&wgClosure, shutdown, ln)
	isReady.Signal()

	for {
		conn, err = ln.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			log.Fatal("[ERROR]", err)
		}

		newConnections <- conn
	}
}

func InitNodeConnections(wg *sync.WaitGroup, myInfos crdt.Infos, newJoinChatCommands <-chan parsestdin.Command, newConnections chan net.Conn, shutdown <-chan struct{}) {
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

			/* Open connection */
			var newConn net.Conn
			newConn, err = openConnection(addr, strconv.Itoa(pt))
			if err != nil {
				log.Println("[ERROR] ", err)
				break
			}

			// init joining process
			_, err = newConn.Write(crdt.NewOperation(crdt.JoinChatByName, chatRoom, myInfos.ToBytes()).ToBytes())
			if err != nil {
				log.Println("[ERROR] ", err)
			}

			newConnections <- newConn
		}
	}
}

func HandleNodes(wg *sync.WaitGroup, newConnections chan net.Conn, toSend <-chan crdt.Operation, toExecute chan<- crdt.Operation, shutdown chan struct{}) {
	var (
		nodes             []node
		connectionsDone   = make(chan uint8, MaxSimultaneousConnections)
		outputConnections = make(chan []byte, MaxMessageSize)
		err               error
	)

	defer func() {
		for _, n := range nodes {
			n.stop()
		}
		wg.Done()
		log.Println("[INFO] HandleNodes stopped")
	}()

	for {
		select {
		case <-shutdown:
			log.Println(" HandleNodes shutting down")
			return

		case newConn := <-newConnections:
			n := newNode(newConn, uint8(len(nodes)+1), outputConnections)
			n.Wg.Add(1)
			go handleConnection(n, connectionsDone)
			nodes = append(nodes, n)

		case slot := <-connectionsDone:
			log.Printf("[INFO] slot %d done\n", slot)
			// TODO build and send operation to chat handler to remove node identified by <slot> from all chats

		case operation := <-toSend:
			slot := operation.GetSlot()
			nodes[slot-1].Input <- operation.ToBytes()

		case operationBytes := <-outputConnections:
			operation := crdt.DecodeOperation(operationBytes[1:])
			if operation.GetOperationType() == crdt.AddNode {
				nodeInfos, _ := crdt.DecodeInfos(operation.GetOperationData())

				// create and saves the new node
				var newConn net.Conn
				newConn, err = openConnection(nodeInfos.GetAddr(), nodeInfos.GetPort())
				if err != nil {
					log.Println("[ERROR] ", err)
				}

				n := newNode(newConn, uint8(len(nodes)+1), outputConnections)
				nodes = append(nodes, n)

				operation.SetSlot(n.slot)
			}

			toExecute <- operation
		}
	}
}

func handleConnection(node node, done chan<- uint8) {
	var (
		wgReadConn      = sync.WaitGroup{}
		shutdown        = make(chan struct{}, 0)
		messageReceived = make(chan []byte, MaxSimultaneousMessages)
	)

	defer func() {
		node.conn.Close()
		wgReadConn.Wait()
		node.Wg.Done()
		done <- node.slot
	}()

	file, _ := node.conn.(*net.TCPConn).File()
	wgReadConn.Add(1)
	go reader.ReadFile(&wgReadConn, file, messageReceived, shutdown)

	for {
		select {
		case <-node.Shutdown:
			return

		case message := <-node.Input:
			// hide slot to receiver
			message[0] = 0
			node.send(message)

		case message, ok := <-messageReceived:
			if !ok {
				// conn closed by the remote client
				return
			}

			// set node slot for chat handler
			message[0] = node.slot
			node.Output <- message
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

func (n *node) stop() {
	close(n.Shutdown)
	n.Wg.Wait()
}

func (n *node) send(message []byte) error {
	_, err := n.conn.Write(message)
	if err != nil {
		return err
	}
	return nil
}

func handleClosure(wg *sync.WaitGroup, shutdown chan struct{}, ln net.Listener) {
	<-shutdown
	err := ln.Close()
	if err != nil {
		log.Fatal(err)
	}

	wg.Done()
}
