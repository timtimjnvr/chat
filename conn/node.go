package conn

import (
	"github/timtimjnvr/chat/reader"
	"net"
	"sync"
)

type (
	node struct {
		slot       uint8
		connection *connection
		Input      chan []byte
		Output     chan []byte
		Wg         *sync.WaitGroup
		Shutdown   chan struct{}
	}
)

func newNode(conn net.Conn, slot uint8, output chan []byte) node {
	return node{
		slot:       slot,
		connection: newConnection(conn),
		Input:      make(chan []byte, MaxSimultaneousMessages),
		Output:     output,
		Wg:         &sync.WaitGroup{},
		Shutdown:   make(chan struct{}, 0),
	}
}

func (n *node) handleNode(done chan<- uint8) {
	var (
		wgReadConn      = sync.WaitGroup{}
		messageReceived = make(chan []byte, MaxSimultaneousMessages)
	)

	defer func() {
		wgReadConn.Wait()
		n.Wg.Done()
		done <- n.slot
	}()

	wgReadConn.Add(1)
	go reader.Read(&wgReadConn, n.connection, messageReceived, n.Shutdown)

	for {
		select {
		case <-n.Shutdown:
			return

		case message := <-n.Input:
			// hide slot to receiver
			message[0] = 0
			n.connection.Write(message)

		case message, ok := <-messageReceived:
			if !ok {
				// connection closed by the remote client
				return
			}

			// set node slot for chat handler
			message[0] = n.slot
			n.Output <- message
		}
	}
}

func (n *node) stop() {
	close(n.Shutdown)
	n.Wg.Wait()
}
