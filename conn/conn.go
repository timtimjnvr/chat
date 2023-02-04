package conn

import (
	"fmt"
	"github.com/pkg/errors"
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

	MaxMessageSize             = 1000
	MaxSimultaneousConnections = 100
	MaxSimultaneousMessages    = 100
)

func HandleNodes(wg *sync.WaitGroup, newConnections chan net.Conn, operationsToSend <-chan []byte, outGoingMessages chan<- []byte, shutdown chan struct{}) {
	var (
		nodes           []node
		connectionsDone = make(chan int, MaxSimultaneousConnections)
	)

	defer func() {
		for _, n := range nodes {
			n.Stop()
		}
		wg.Done()
		log.Println("[INFO] HandleNodes stopped")
	}()

	select {
	case <-shutdown:
		return

	case newConn := <-newConnections:
		newNode := NewNode(newConn)
		newNode.Wg.Add(1)
		go handleConnection(newNode, outGoingMessages, connectionsDone)
		nodes = append(nodes, newNode)

	case slot := <-connectionsDone:
		log.Printf("[INFO] slot %d done\n", slot)
		// TODO
		/*
			build and send operation to chat handler to remove node identified by <slot> from all chats
		*/

	case <-operationsToSend:
		// TODO
		/*
			get node slot
			sends to right node
		*/
	}
}

func ListenAndServe(wg *sync.WaitGroup, addr, port string, newConnections chan net.Conn, shutdown chan struct{}) {
	var (
		conn      net.Conn
		wgClosure = sync.WaitGroup{}
		err       error
	)

	defer func() {
		wgClosure.Wait()
		wg.Done()
		log.Println("[INFO] listenAndServe stopped")
	}()

	ln, err := net.Listen(transportProtocol, fmt.Sprintf("%s:%s", addr, port))
	if err != nil {
		log.Fatal(err)
	}

	wgClosure.Add(1)
	go handleClosure(&wgClosure, shutdown, ln)
	for {
		log.Println(fmt.Sprintf("[INFO] Accepting connections on port %s", port))
		conn, err = ln.Accept()
		if err != nil && errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			log.Println("[WARNING] err Accept :", err)
			continue
		}

		newConnections <- conn
	}
}

func handleConnection(node node, outGoingMessages chan<- []byte, done chan<- int) {
	log.Println("[INFO] new conn")

	var (
		wgReadConn      = sync.WaitGroup{}
		messageReceived = make(chan []byte, MaxSimultaneousMessages)
	)

	defer func() {
		node.Conn.Close()
		wgReadConn.Wait()
		node.Wg.Done()
		done <- node.slot
		log.Println("[INFO] conn lost for ", node.Conn.LocalAddr())
	}()

	wgReadConn.Add(1)
	go readConn(&wgReadConn, node.Conn, messageReceived, node.Shutdown)

	for {
		select {
		case <-node.Shutdown:
			return

		case message, ok := <-messageReceived:
			if !ok {
				// conn closed on the other side
				return
			}

			// add slot to message
			outGoingMessages <- append([]byte{uint8(node.slot)}, message...)
		}
	}
}

func OpenConnection(ip string, port int) (net.Conn, error) {
	if ip == localhost || ip == localhostDecimalPointed {
		ip = ""
	}

	conn, err := net.Dial(transportProtocol, fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func NewNode(conn net.Conn) node {
	return node{
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

func readConn(wg *sync.WaitGroup, conn net.Conn, messages chan []byte, shutdown chan struct{}) {
	defer func() {
		close(messages)
		wg.Done()
		log.Println("[INFO] readConn stopped")
	}()

	for {
		select {
		case <-shutdown:
			return

		default:
			buffer := make([]byte, MaxMessageSize)
			n, err := conn.Read(buffer)
			if err != nil {
				return
			}
			if n > 0 {
				log.Print(string(buffer))
				messages <- buffer[0:n]
			}
		}
	}
}

func handleClosure(wg *sync.WaitGroup, shutdown chan struct{}, ln net.Listener) {
	<-shutdown
	err := ln.Close()
	if err != nil {
		log.Fatal(err)
	}

	wg.Done()
	log.Println("[INFO] handleClosure stopped")
}
