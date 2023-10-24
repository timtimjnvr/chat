package conn

import (
	"fmt"
	"github/timtimjnvr/chat/crdt"
	"github/timtimjnvr/chat/reader"
	"log"
	"net"
	"sync"
)

type (
	slot uint8

	node struct {
		slot slot
		conn *conn

		Input  chan []byte
		Output chan<- []byte

		Wg       *sync.WaitGroup
		Shutdown chan struct{}
	}

	NodeHandler struct {
		nodes map[slot]*node

		Wg       *sync.WaitGroup
		Shutdown chan struct{}
	}
)

func newNode(conn net.Conn, slot slot, output chan<- []byte) (*node, error) {
	c, err := newConn(conn)
	if err != nil {
		return nil, err
	}

	return &node{
		slot:     slot,
		conn:     c,
		Input:    make(chan []byte),
		Output:   output,
		Wg:       &sync.WaitGroup{},
		Shutdown: make(chan struct{}, 0),
	}, nil
}

func (n *node) start(done chan<- slot) {
	outputConnection := make(chan []byte)

	defer func() {
		n.Wg.Done()
		done <- n.slot
	}()

	isDone := make(chan struct{})
	go reader.Read(n.conn, outputConnection, reader.Separator, n.Shutdown, isDone)

	for {
		select {
		case <-isDone:
			return

		case <-n.Shutdown:
			return

		case message := <-n.Input:
			// hide own slot to remote client
			message = resetSlot(message)
			n, err := n.conn.Write(message)
			if err != nil {
				fmt.Printf("Write %d, %s\n", n, err)
			}

		case message, ok := <-outputConnection:
			if !ok {
				// conn closed by the remote client
				return
			}

			// set node slot for chat NodeHandler
			n.setSlot(message)
			n.Output <- message
		}
	}
}

func (n *node) setSlot(message []byte) []byte {
	message[0] = uint8(n.slot)
	return message
}

func (n *node) stop() {
	if n == nil {
		return
	}

	close(n.Shutdown)
	n.Wg.Wait()
}

func (d *NodeHandler) getNextSlot() slot {
	length := len(d.nodes)
	for s, n := range d.nodes {
		if n == nil {
			return s
		}
	}

	return slot(length + 1)
}

func NewNodeHandler(shutdown chan struct{}) *NodeHandler {
	return &NodeHandler{
		nodes:    make(map[slot]*node),
		Wg:       &sync.WaitGroup{},
		Shutdown: shutdown,
	}
}

func (d *NodeHandler) Start(newConnections <-chan net.Conn, toSend <-chan *crdt.Operation, toExecute chan<- *crdt.Operation) {

	var (
		done        = make(chan slot)
		outputNodes = make(chan []byte)
	)

	defer func() {
		for _, n := range d.nodes {
			n.stop()
		}

		d.Wg.Done()
	}()

	for {
		select {
		case <-d.Shutdown:
			return

		case c := <-newConnections:
			s := d.getNextSlot()
			n, err := newNode(c, d.getNextSlot(), outputNodes)
			if err != nil {
				log.Println("[ERROR] ", err)
				continue
			}

			n.Wg.Add(1)
			go n.start(done)
			d.nodes[s] = n

		case s := <-done:
			quitOperation := crdt.NewOperation(crdt.Quit, "", nil)
			quitOperation.Slot = uint8(s)
			toExecute <- quitOperation
			d.nodes[s] = nil

		case operation := <-toSend:
			s := slot(operation.Slot)
			if n, exist := d.nodes[s]; exist {
				n.Input <- operation.ToBytes()
				if operation.Typology == crdt.KillNode {
					n.stop()
				}
			}

		case operationBytes := <-outputNodes:
			operation, err := crdt.DecodeOperation(operationBytes)
			if err != nil {
				log.Println("[ERROR] ", err)
				continue
			}

			// need to create connection and set slot in operation
			if operation.Typology == crdt.AddNode {
				newNodeInfos, ok := operation.Data.(*crdt.NodeInfos)
				if !ok {
					log.Println("[ERROR] can't parse op data to NodeInfos")
					continue
				}

				// establish connection and set slot
				var c net.Conn
				c, err = openConnection(newNodeInfos.Address, newNodeInfos.Port)
				if err != nil {
					log.Println("[ERROR] ", err)
					break
				}

				s := d.getNextSlot()
				n, err := newNode(c, d.getNextSlot(), outputNodes)
				if err != nil {
					log.Println("[ERROR] ", err)
					continue
				}

				n.Wg.Add(1)
				go n.start(done)
				d.nodes[s] = n

				operation.Slot = uint8(s)
			}

			toExecute <- operation
		}
	}
}

func resetSlot(message []byte) []byte {
	message[0] = 0
	return message
}
