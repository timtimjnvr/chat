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
		newConnections  = make(chan net.Conn, MaxSimultaneousConnections)
		toSend          = make(chan crdt.Operation, MaxSimultaneousMessages)
		toExecute       = make(chan crdt.Operation, MaxSimultaneousMessages)
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

	// wait for conn1 to be handled and saved
	<-time.Tick(10 * time.Millisecond)

	// killing conn1 by closing conn2
	conn2.Close()

	timeout := time.Tick(maxTestDuration)
	quitOperation := crdt.NewOperation(crdt.Quit, "", []byte{})
	quitOperation.SetSlot(1)
	expectedMessage := quitOperation.ToBytes()

	select {
	case <-timeout:
		assert.Fail(t, "test timeout")
		return
	case op := <-toExecute:
		assert.Equal(t, op.ToBytes(), expectedMessage, "did not received expected quit operation")
	}
}
