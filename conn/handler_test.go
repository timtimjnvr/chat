package conn

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github/timtimjnvr/chat/crdt"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDriver_StartAndStop(t *testing.T) {
	var (
		output          = make(chan []byte, maxMessageSize)
		done            = make(chan slot, 2)
		maxTestDuration = 1 * time.Second
	)

	connReader, connSender, err := helperGetConnections("12349")
	if err != nil {
		assert.Fail(t, "failed to create a conn")
	}

	reader, err := newNode(connReader, 1, output)
	if err != nil {
		assert.Fail(t, "failed to create node")
	}
	reader.Wg.Add(1)
	go reader.start(done)

	sender, err := newNode(connSender, 1, output)
	if err != nil {
		assert.Fail(t, "failed to create node")
	}

	sender.Wg.Add(1)
	go sender.start(done)

	var (
		message         = []byte{2, 1, 2, 3, 4, 5} // slot set to node slot sender
		expectedMessage = []byte{1, 1, 2, 3, 4, 5} // slot set to node slot receiver
	)

	sender.Input <- message

	timeout := time.Tick(maxTestDuration)
	select {
	case <-timeout:
		assert.Fail(t, "test timeout")
	case received := <-output:
		assert.Equal(t, expectedMessage, received, "messages sent and received are not equal")
	}

	// test closure (on both side sender & receiver) -> reception of done message
	sender.stop()
	defer func() {
		sender.Wg.Wait()
		reader.Wg.Wait()
	}()

	timeout = time.Tick(maxTestDuration)
	numberOfDone := 0

	for {
		select {
		case <-timeout:
			assert.Fail(t, "test timeout")
			return
		case <-done:
			numberOfDone++
			if numberOfDone == 2 {
				return
			}
		}
	}
}

func TestDriver_StartStopNodesAndSendQuit(t *testing.T) {
	var (
		maxTestDuration = 2 * time.Second
		shutdown        = make(chan struct{}, 0)
		nh              = NewNodeHandler(shutdown)
		newConnections  = make(chan net.Conn)
		toSend          = make(chan crdt.Operation)
		toExecute       = make(chan crdt.Operation)
	)

	defer func() {
		close(shutdown)
		nh.Wg.Wait()
	}()

	nh.Wg.Add(1)
	go nh.Start(newConnections, toSend, toExecute)

	conn1, conn2, err := helperGetConnections("12346")
	if err != nil {
		assert.Fail(t, "failed to create a conn")
	}

	newConnections <- conn1

	// killing conn1 by closing conn2
	conn2.Close()

	expectedQuitOperation := crdt.NewOperation(crdt.Quit, "", []byte{})
	expectedQuitOperation.SetSlot(1)
	expectedBytes := expectedQuitOperation.ToBytes()

	timeout := time.Tick(maxTestDuration)
	select {
	case <-timeout:
		assert.Fail(t, "test timeout")
		return
	case op := <-toExecute:
		assert.Equal(t, op.ToBytes(), expectedBytes, "did not received expected quit operation")
	}
}

func TestNodeHandler_Send(t *testing.T) {
	// creating linked connections
	conn1, conn2, err := helperGetConnections("12347")
	if err != nil {
		assert.Fail(t, "failed to create a conn")
	}

	var (
		output = make(chan []byte, maxMessageSize)
		done   = make(chan slot, 1)
	)

	nodeReader, err := newNode(conn2, 0, output)
	if err != nil {
		assert.Fail(t, "failed to create node")
	}
	nodeReader.Wg.Add(1)
	go nodeReader.start(done)

	var (
		maxTestDuration = 1 * time.Second
		shutdown        = make(chan struct{}, 0)
		nh              = NewNodeHandler(shutdown)
		newConnections  = make(chan net.Conn)
		toSend          = make(chan crdt.Operation)
		toExecute       = make(chan crdt.Operation)
	)

	defer func() {
		close(shutdown)
		nh.Wg.Wait()
		nodeReader.stop()
	}()

	nh.Wg.Add(1)
	go nh.Start(newConnections, toSend, toExecute)

	newConnections <- conn1
	messageOperation := crdt.NewOperation(crdt.AddMessage, "test-chat", []byte("I love Unit Testing"))
	messageOperation.SetSlot(1)

	expectedMessageOperation := crdt.NewOperation(crdt.AddMessage, "test-chat", []byte("I love Unit Testing"))
	expectedMessageOperation.SetSlot(0)
	expectedBytes := expectedMessageOperation.ToBytes()

	toSend <- messageOperation

	timeout := time.Tick(maxTestDuration)
	select {
	case <-timeout:
		assert.Fail(t, "test timeout")
		return
	case m := <-output:
		assert.Equal(t, m, expectedBytes, "did not received expected operation bytes")
	}
}

func TestNodeHandler_SOMAXCONNNodesStartAndStop(t *testing.T) {
	var (
		maxTestDuration = 3 * time.Second
		shutdown        = make(chan struct{}, 0)
		nh              = NewNodeHandler(shutdown)
		newConnections  = make(chan net.Conn)
		toSend          = make(chan crdt.Operation)
		toExecute       = make(chan crdt.Operation)

		firstPort  = 1235
		maxNode    = syscall.SOMAXCONN
		connSaving = make(map[int]net.Conn, maxNode)
	)
	nh.Wg.Add(1)
	go nh.Start(newConnections, toSend, toExecute)
	defer nh.Wg.Wait()

	for i := 0; i < maxNode; i++ {
		conn1, conn2, err := helperGetConnections(fmt.Sprintf("%d", firstPort))
		if err != nil {
			assert.Fail(t, "failed to create a conn")
		}

		connSaving[i] = conn2
		newConnections <- conn1
		firstPort++
	}

	expectedQuitOperation := crdt.NewOperation(crdt.Quit, "", []byte{})

	// killing all connections and checking messages
	for i := 0; i < maxNode; i++ {
		timeout := time.Tick(maxTestDuration)
		connSaving[i].Close()
		expectedQuitOperation.SetSlot(uint8(i + 1))
		expectedBytes := expectedQuitOperation.ToBytes()

		select {
		case <-timeout:
			assert.Fail(t, "test timeout")
			return
		case op := <-toExecute:
			assert.Equal(t, op.ToBytes(), expectedBytes, "did not received expected quit operation")
		}
	}
}
