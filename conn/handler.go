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
	// slot identifies a TCP connection in one node referential (it get its values between 0 and 255).
	// Any slot between 1 and 255 identifies an active TCP connection in the node handler.
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
		debugMode bool
		nodes     map[slot]*node

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

func (d *NodeHandler) SetDebugMode() {
	d.debugMode = true
}

func (d *NodeHandler) Start(newConnections <-chan net.Conn, toSend <-chan *crdt.Operation, toExecute chan<- *crdt.Operation) {
	var (
		nodeAccess  = &sync.Mutex{}
		done        = make(chan slot)
		outputNodes = make(chan []byte)
	)

	defer func() {
		for _, n := range d.nodes {
			n.stop()
		}

		d.Wg.Done()
	}()

	d.Wg.Add(1)
	go func() {
		defer d.Wg.Done()

		for {
			select {
			case <-d.Shutdown:
				return

			case operation := <-toSend:
				if d.debugMode {
					fmt.Println("[DEBUG] node Handler", crdt.GetOperationName(operation.Typology), "operation to send")
				}

				// Set slot
				s := slot(operation.Slot)
				nodeAccess.Lock()
				if n, exist := d.nodes[s]; exist && n != nil {
					n.Input <- operation.ToBytes()
					if operation.Typology == crdt.KillNode {
						n.stop()
					}
				}
				nodeAccess.Unlock()
			}
		}
	}()

	for {
		select {
		case <-d.Shutdown:
			if d.debugMode {
				fmt.Println("[DEBUG] node Handler shuting down")
			}

			return

		case c := <-newConnections:
			s := d.getNextSlot()
			if d.debugMode {
				fmt.Println("[DEBUG] node Handler", "New connection", s)
			}

			n, err := newNode(c, d.getNextSlot(), outputNodes)
			if err != nil {
				log.Println("[ERROR] ", err)
				continue
			}

			n.Wg.Add(1)
			go n.start(done)
			nodeAccess.Lock()
			d.nodes[s] = n
			nodeAccess.Unlock()

		case s := <-done:
			if d.debugMode {
				fmt.Println("[DEBUG] node Handler", "Node done", s)
			}

			quitOperation := crdt.NewOperation(crdt.KillNode, "", nil)
			quitOperation.Slot = uint8(s)
			toExecute <- quitOperation
			nodeAccess.Lock()
			d.nodes[s] = nil
			nodeAccess.Unlock()

		case operationBytes := <-outputNodes:
			if d.debugMode {
				fmt.Println("[DEBUG] node Handler received operationBytes")
			}

			operation, err := crdt.DecodeOperation(operationBytes)
			if err != nil {
				log.Println("[ERROR] ", err)
				continue
			}

			if d.debugMode {
				fmt.Println("[DEBUG] node Handler", crdt.GetOperationName(operation.Typology), "operation received")
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
				if d.debugMode {
					fmt.Println("[DEBUG] node Handler establishing connection")
				}

				c, err = openConnection(newNodeInfos.Address, newNodeInfos.Port)
				if err != nil {
					log.Println("[ERROR] ", err)
					break
				}

				if d.debugMode {
					fmt.Println("[DEBUG] node Handler connection established")
				}

				s := d.getNextSlot()
				n, err := newNode(c, d.getNextSlot(), outputNodes)
				if err != nil {
					log.Println("[ERROR] ", err)
					continue
				}

				n.Wg.Add(1)
				go n.start(done)
				nodeAccess.Lock()
				d.nodes[s] = n
				nodeAccess.Unlock()

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
