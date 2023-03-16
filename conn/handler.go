package conn

import (
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

	driver struct {
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
	var (
		wgReadConn       = sync.WaitGroup{}
		outputConnection = make(chan []byte)
	)

	defer func() {
		wgReadConn.Wait()
		n.Wg.Done()
		done <- n.slot
	}()

	wgReadConn.Add(1)
	go reader.Read(&wgReadConn, n.conn, outputConnection, reader.Separator, n.Shutdown)

	for {
		select {
		case <-n.Shutdown:
			return

		case message := <-n.Input:
			// hide own slot to remote client
			message = resetSlot(message)
			n.conn.Write(message)

		case message, ok := <-outputConnection:
			if !ok {
				// conn closed by the remote client
				return
			}

			// set node slot for chat driver
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

func (d *driver) getNextSlot() slot {
	length := len(d.nodes)
	for s, n := range d.nodes {
		if n == nil {
			return s
		}
	}

	return slot(length + 1)
}

func NewNodeDriver(shutdown chan struct{}) *driver {
	return &driver{
		nodes:    make(map[slot]*node),
		Wg:       &sync.WaitGroup{},
		Shutdown: shutdown,
	}
}

func (d *driver) Start(newConnections <-chan net.Conn, toSend <-chan crdt.Operation, toExecute chan<- crdt.Operation) {
	d.Wg.Add(1)

	var (
		done        = make(chan slot)
		outputNodes = make(chan []byte)
		err         error
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
			var n *node
			n, err = newNode(c, d.getNextSlot(), outputNodes)
			if err != nil {
				log.Println("[ERROR] ", err)
				continue
			}

			n.Wg.Add(1)
			go n.start(done)
			d.nodes[s] = n

		case s := <-done:
			quitOperation := crdt.NewOperation(crdt.Quit, "", []byte{})
			quitOperation.SetSlot(uint8(s))
			toExecute <- quitOperation
			d.nodes[s] = nil

		case operation := <-toSend:
			s := slot(operation.GetSlot())
			if n, exist := d.nodes[s]; exist {
				n.Input <- operation.ToBytes()
			}

		case operationBytes := <-outputNodes:
			operation := crdt.DecodeOperation(operationBytes)
			toExecute <- operation
		}
	}
}

func resetSlot(message []byte) []byte {
	message[0] = 0
	return message
}
