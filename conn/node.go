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
		slot       slot
		connection *connection
		Input      chan []byte
		Output     chan []byte
		Wg         *sync.WaitGroup
		Shutdown   chan struct{}
	}

	nodeHandler struct {
		nodes    map[slot]*node
		Wg       *sync.WaitGroup
		Shutdown chan struct{}
	}
)

func newNode(conn net.Conn, slot slot, output chan []byte) (node, error) {
	c, err := newConnection(conn)
	if err != nil {
		return node{}, err
	}
	return node{
		slot:       slot,
		connection: c,
		Input:      make(chan []byte, MaxSimultaneousMessages),
		Output:     output,
		Wg:         &sync.WaitGroup{},
		Shutdown:   make(chan struct{}, 0),
	}, nil
}

func (n *node) start(done chan<- slot) {
	n.Wg.Add(1)
	var (
		wgReadConn       = sync.WaitGroup{}
		outputConnection = make(chan []byte, MaxSimultaneousMessages)
	)

	defer func() {
		wgReadConn.Wait()
		n.Wg.Done()
		done <- n.slot
	}()

	wgReadConn.Add(1)
	go reader.Read(&wgReadConn, n.connection, outputConnection, reader.Separator, n.Shutdown)

	for {
		select {
		case <-n.Shutdown:
			return

		case message := <-n.Input:
			// hide slot to receiver
			message = resetSlot(message)
			n.connection.Write(message)

		case message, ok := <-outputConnection:
			if !ok {
				// connection closed by the remote client
				return
			}

			// set node slot for chat handler
			n.setSlot(message)
			n.Output <- message
		}
	}
}

func resetSlot(message []byte) []byte {
	message[0] = 0
	return message
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

func (nh *nodeHandler) getNextSlot() slot {
	length := len(nh.nodes)
	for s, n := range nh.nodes {
		if n == nil {
			return s
		}
	}

	return slot(length + 1)
}

func NewNodeHandler(shutdown chan struct{}) *nodeHandler {
	return &nodeHandler{
		nodes:    make(map[slot]*node),
		Wg:       &sync.WaitGroup{},
		Shutdown: shutdown,
	}
}
func (nh *nodeHandler) Start(newConnections chan net.Conn, toSend <-chan crdt.Operation, toExecute chan<- crdt.Operation) {
	nh.Wg.Add(1)

	var (
		done        = make(chan slot, MaxSimultaneousConnections)
		outputNodes = make(chan []byte, MaxSimultaneousMessages)
		err         error
	)

	defer func() {
		for _, n := range nh.nodes {
			n.stop()
		}

		nh.Wg.Done()
		log.Println("[INFO] nodeHandler stopped")
	}()

	for {
		select {
		case <-nh.Shutdown:
			log.Println("[INFO] nodeHandler shutting down")
			return

		case newConn := <-newConnections:
			s := nh.getNextSlot()
			var n node
			n, err = newNode(newConn, nh.getNextSlot(), outputNodes)
			if err != nil {
				log.Println("[ERROR] ", err)
				continue
			}

			go n.start(done)
			nh.nodes[s] = &n

		case s := <-done:
			log.Printf("[INFO] slot %d done\n", s)
			quitOperation := crdt.NewOperation(crdt.Quit, "", []byte{})
			quitOperation.SetSlot(uint8(s))
			toExecute <- quitOperation
			nh.nodes[s] = nil

		case operation := <-toSend:
			s := slot(operation.GetSlot())
			if n, exist := nh.nodes[s]; exist {
				n.Input <- operation.ToBytes()
			}

		case operationBytes := <-outputNodes:
			operation := crdt.DecodeOperation(operationBytes[1:])

			if operation.GetOperationType() == crdt.AddNode {
				nodeInfos, _ := crdt.DecodeInfos(operation.GetOperationData())

				// create and saves the new node
				var newConn net.Conn
				newConn, err = openConnection(nodeInfos.GetAddr(), nodeInfos.GetPort())
				if err != nil {
					log.Println("[ERROR] ", err)
				}
				s := nh.getNextSlot()
				n, err := newNode(newConn, s, outputNodes)
				if err != nil {
					log.Println("[ERROR] ", err)
					continue
				}

				nh.nodes[s] = &n

				operation.SetSlot(uint8(s))
			}

			toExecute <- operation
		}
	}
}
