package conn

import (
	"fmt"
	"github/timtimjnvr/chat/crdt"
	"github/timtimjnvr/chat/reader"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
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

		Wg *sync.WaitGroup
	}

	NodeHandler struct {
		nodeStorage NodeStorage
		nodes       map[slot]*node

		Wg *sync.WaitGroup
	}

	NodeStorage interface {
		GetNodeBySlot(slot uint8) (*crdt.NodeInfos, error)
	}
)

func newNode(conn net.Conn, slot slot, output chan<- []byte) (*node, error) {
	c, err := newConn(conn)
	if err != nil {
		return nil, err
	}

	return &node{
		slot:   slot,
		conn:   c,
		Input:  make(chan []byte),
		Output: output,
		Wg:     &sync.WaitGroup{},
	}, nil
}

func (n *node) start(done chan<- slot) {
	var (
		outputConnection = make(chan []byte)
		stopReading      = make(chan struct{})
		isClosing        = atomic.Bool{}
	)
	defer func() {
		isClosing.Store(true)
		close(stopReading)
		n.Wg.Done()
	}()

	go reader.Read(n.conn, outputConnection, reader.Separator, stopReading)

	for {
		select {
		case message, more := <-n.Input:
			if !more {
				return
			}

			// Hide own slot to remote client
			message = resetSlot(message)
			_, err := n.conn.Write(message)
			if err != nil {
				fmt.Printf("Write: %s\n", err)
				// TCP connection need to be re established
				done <- n.slot
				return
			}

		case message, more := <-outputConnection:
			if !more {
				// TCP connection closed and need to be re established
				if !isClosing.Load() {
					done <- n.slot
					return
				}
			}

			// Set node slot for chat NodeHandler
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
	close(n.Input)
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

func NewNodeHandler(nodeStorage NodeStorage) *NodeHandler {
	return &NodeHandler{
		nodeStorage: nodeStorage,
		nodes:       make(map[slot]*node),
		Wg:          &sync.WaitGroup{},
	}
}

func (d *NodeHandler) Start(newConnections <-chan net.Conn, toSend <-chan *crdt.Operation, toExecute chan<- *crdt.Operation) {
	var (
		nodeAccess                 = &sync.Mutex{}
		done                       = make(chan slot)
		outputNodes                = make(chan []byte)
		stopTCPConnectionsHandling = make(chan struct{}, 0)
		TCPHandling                = &sync.WaitGroup{}
	)

	defer func() {
		TCPHandling.Done()
		close(toExecute)
		// Hugly but needed
		// this is the last go routine to exit
		// when this exits all leaving TCP connections will be closed
		// we want them to be closed cleanly by the remote nodes
		// wen the KillNode operation is received
		<-time.Tick(5 * time.Second)
		d.Wg.Done()
	}()

	d.Wg.Add(1)
	go func() {
		defer d.Wg.Done()

		for {
			select {
			case operation, more := <-toSend:
				// Received shutdown
				if !more {
					close(stopTCPConnectionsHandling)
					TCPHandling.Wait()
					return
				}

				// Set slot
				s := slot(operation.Slot)

				nodeAccess.Lock()
				// Broadcast
				if s == 0 {
					for _, n := range d.nodes {
						if n != nil {
							n.Input <- operation.ToBytes()
							// TODO find a way to wait node
						}
					}
				} else {
					if n, exist := d.nodes[s]; exist && n != nil {
						n.Input <- operation.ToBytes()
						// TODO find a way to wait node
					}
				}

				nodeAccess.Unlock()
			}
		}
	}()

	TCPHandling.Add(1)
	for {
		select {
		case <-stopTCPConnectionsHandling:

			return

		case c := <-newConnections:
			fmt.Println("[DEBUG] node Handler", "New connection")
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

			// TCP connection closed unexpectedly
		case s := <-done:

			nodeInfos, err := d.nodeStorage.GetNodeBySlot(uint8(s))
			if err != nil {
				log.Println("[ERROR] ", err)
				continue
			}

			c, err := openConnection(nodeInfos.Address, nodeInfos.Port)
			if err != nil {
				log.Println("[ERROR] ", err)
				continue
			}

			resetNode, err := newNode(c, s, outputNodes)
			if err != nil {
				log.Println("[ERROR] ", err)
				continue
			}

			nodeAccess.Lock()
			d.nodes[s] = resetNode
			nodeAccess.Unlock()

		case operationBytes := <-outputNodes:

			operation, err := crdt.DecodeOperation(operationBytes)
			if err != nil {
				log.Println("[ERROR] ", err)
				continue
			}

			// Open TCP connection
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
				nodeAccess.Lock()
				d.nodes[s] = n
				nodeAccess.Unlock()

				operation.Slot = uint8(s)
			}

			// Close TCP connection
			if operation.Typology == crdt.KillNode {
				nodeAccess.Lock()

				// Kill all TCP connections
				if operation.Slot == 0 {
					for s, n := range d.nodes {
						if n != nil {
							fmt.Println("closing connection")
							n.stop()
							d.nodes[s] = nil
						}
					}
				} else {
					// Kill specific TCP connection
					if n, exists := d.nodes[slot(operation.Slot)]; exists && n != nil {
						n.stop()
						d.nodes[slot(operation.Slot)] = nil
					}
				}

				nodeAccess.Unlock()
			}

			toExecute <- operation
		}
	}
}

func resetSlot(message []byte) []byte {
	message[0] = 0
	return message
}
