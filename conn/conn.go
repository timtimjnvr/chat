package conn

import (
	"chat/node"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"log"
	"net"
	"sync"
)

const (
	maxMessagesInConnection = 100
	MaxMessageSize          = 10000
)

func ListenAndServe(wg *sync.WaitGroup, newNodes chan *node.Node, shutdown chan struct{}, transportProtocol, addr, port string) {
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

		newNode := node.NewNode(conn)
		newNodes <- newNode
	}
}

func readConn(wg *sync.WaitGroup, conn net.Conn, messages chan string, shutdown chan struct{}) {
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
				messages <- string(buffer[0:n])
			}
		}
	}
}

func HandleConnection(node *node.Node, done chan<- uuid.UUID, shutdown chan struct{}) {
	log.Println("[INFO] new conn")

	var wgReadConn = sync.WaitGroup{}

	defer func() {
		node.Business.Conn.Close()
		wgReadConn.Wait()
		node.Business.Wg.Done()
		done <- node.Infos.Id
		log.Println("[INFO] conn lost for ", node.Business.Conn.LocalAddr())
	}()

	wgReadConn.Add(1)
	go readConn(&wgReadConn, node.Business.Conn, node.Business.MessagesReceived, shutdown)

	for {
		select {
		case <-shutdown:
			return

		case _, ok := <-node.Business.MessagesReceived:
			if !ok {
				// conn closed on the other side
				return
			}
		}
	}
}

func OpenConnection(protocol, ip string, port int) (net.Conn, error) {
	conn, err := net.Dial(protocol, fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return nil, err
	}

	return conn, nil
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
	log.Println("[INFO] handleClosure stopped")
}
