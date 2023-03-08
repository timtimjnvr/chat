package conn

import (
	"github/timtimjnvr/chat/crdt"
	"github/timtimjnvr/chat/reader"
	"log"
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
	c, _ := newConnection(conn)
	return node{
		slot:       slot,
		connection: c,
		Input:      make(chan []byte, MaxSimultaneousMessages),
		Output:     output,
		Wg:         &sync.WaitGroup{},
		Shutdown:   make(chan struct{}, 0),
	}
}

func (n *node) start(done chan<- uint8) {
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
	go reader.Read(&wgReadConn, n.connection, messageReceived, reader.Separator, n.Shutdown)

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
			go n.start(connectionsDone)
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
