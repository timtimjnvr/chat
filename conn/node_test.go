package conn

import (
	"github.com/stretchr/testify/assert"
	"github/timtimjnvr/chat/crdt"
	"net"
	"testing"
	"time"
)

func TestNode_StartAndStop(t *testing.T) {
	var (
		output          = make(chan []byte, MaxMessageSize)
		done            = make(chan slot, 2)
		maxTestDuration = 1 * time.Second
	)

	connReader, connSender, err := helperGetConnections("12349")
	if err != nil {
		assert.Fail(t, "failed to create a connection")
	}

	reader, err := newNode(connReader, 1, output)
	if err != nil {
		assert.Fail(t, "failed to create node")
	}
	go reader.start(done)

	sender, err := newNode(connSender, 1, output)
	if err != nil {
		assert.Fail(t, "failed to create node")
	}
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

func TestNodeHandler_StartStopNodeAndSendQuit(t *testing.T) {
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

	go nh.Start(newConnections, toSend, toExecute)

	conn1, conn2, err := helperGetConnections("12346")
	if err != nil {
		assert.Fail(t, "failed to create a connection")
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
		assert.Fail(t, "failed to create a connection")
	}

	var (
		output = make(chan []byte, MaxMessageSize)
		done   = make(chan slot, 1)
	)

	nodeReader, err := newNode(conn2, 0, output)
	if err != nil {
		assert.Fail(t, "failed to create node")
	}

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
