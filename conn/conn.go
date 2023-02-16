package conn

import (
	"fmt"
	"github.com/pkg/errors"
	"github/timtimjnvr/chat/crdt"
	"github/timtimjnvr/chat/reader"
	"log"
	"net"
	"sync"
)

type (
	node struct {
		slot     int
		Conn     net.Conn
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
)

func HandleNodes(wg *sync.WaitGroup, newConnections chan net.Conn, operationsToSend <-chan []byte, outGoingOperations chan<- []byte, shutdown chan struct{}) {
	var (
		nodes           []node
		connectionsDone = make(chan int, MaxSimultaneousConnections)
		fromConnections = make(chan []byte)
		err             error
	)

	defer func() {
		for _, n := range nodes {
			n.Stop()
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
			newNode := NewNode(newConn, len(nodes))
			newNode.Wg.Add(1)
			go handleConnection(newNode, fromConnections, connectionsDone)
			nodes = append(nodes, newNode)

		case slot := <-connectionsDone:
			log.Printf("[INFO] slot %d done\n", slot)
			// TODO
			/*
				build and send operationBytes to chat handler to remove node identified by <slot> from all chats
			*/

		case operationBytes := <-operationsToSend:
			slot := operationBytes[0]
			_ = Send(nodes[slot].Conn, operationBytes[:1])

		case operationBytes := <-fromConnections:
			operation := crdt.DecodeOperation(operationBytes[1:])

			if operation.GetOperationType() == crdt.AddNode {
				nodeInfos, _ := crdt.DecodeInfos(operation.GetOperationData())

				// create and saves the new node
				var newConn net.Conn
				newConn, err = OpenConnection(nodeInfos.GetAddr(), nodeInfos.GetPort())
				if err != nil {
					log.Println("[ERROR] ", err)
				}

				newNode := NewNode(newConn, len(nodes))
				nodes = append(nodes, newNode)

				operationBytes = crdt.AddSlot(newNode.slot, operation.ToBytes())
			}

			outGoingOperations <- operationBytes
		}
	}
}

func ListenAndServe(wg *sync.WaitGroup, isListening *sync.Cond, addr, port string, newConnections chan net.Conn, shutdown chan struct{}) {
	var (
		conn      net.Conn
		wgClosure = sync.WaitGroup{}
		err       error
	)

	defer func() {
		close(newConnections)
		wgClosure.Wait()
		wg.Done()
	}()

	ln, err := net.Listen(transportProtocol, fmt.Sprintf("%s:%s", addr, port))
	if err != nil {
		log.Println("[ERROR]", err)
		return
	}

	wgClosure.Add(1)
	go handleClosure(&wgClosure, shutdown, ln)

	isListening.Signal()

	for {
		conn, err = ln.Accept()
		if err != nil && errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			log.Fatal(err)
		}

		newConnections <- conn
	}
}

func handleConnection(node node, outGoingMessages chan<- []byte, done chan<- int) {
	log.Println("[INFO] new conn")

	var (
		wgReadConn      = sync.WaitGroup{}
		shutdown        = make(chan struct{}, 0)
		messageReceived = make(chan []byte, MaxSimultaneousMessages)
	)

	defer func() {
		node.Conn.Close()
		wgReadConn.Wait()
		node.Wg.Done()
		done <- node.slot
		log.Println("[INFO] conn lost for ", node.Conn.LocalAddr())
	}()
	file, _ := node.Conn.(*net.TCPConn).File()
	wgReadConn.Add(1)
	go reader.ReadFile(&wgReadConn, file, messageReceived, shutdown)

	for {
		select {
		case <-node.Shutdown:
			return

		case message, ok := <-messageReceived:
			if !ok {
				log.Println("not ok")
				// conn closed on the other side
				return
			}
			log.Println("message received :", string(message))
			// add slot to message
			outGoingMessages <- append([]byte{uint8(node.slot)}, message...)
		}
	}
}

func OpenConnection(ip string, port string) (net.Conn, error) {
	if ip == localhost || ip == localhostDecimalPointed {
		ip = ""
	}

	conn, err := net.Dial(transportProtocol, fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func NewNode(conn net.Conn, slot int) node {
	return node{
		slot:     slot,
		Conn:     conn,
		Wg:       &sync.WaitGroup{},
		Shutdown: make(chan struct{}, 0),
	}
}

func (n *node) Stop() {
	close(n.Shutdown)
	n.Wg.Wait()
}

func Send(conn net.Conn, message []byte) error {
	_, err := conn.Write(message)
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
