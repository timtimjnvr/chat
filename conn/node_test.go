package conn

import (
	"github.com/stretchr/testify/assert"
	"net"
	"sync"
	"testing"
	"time"
)

func TestHandleConnection(t *testing.T) {
	// TODO : test message sending (received by receiver)
	var (
		wgListen       = sync.WaitGroup{}
		shutdown       = make(chan struct{}, 0)
		lock           = sync.Mutex{}
		isListening    = sync.NewCond(&lock)
		newConnections = make(chan net.Conn, MaxSimultaneousConnections)
		output         = make(chan []byte, MaxMessageSize)
		done           = make(chan uint8, 2)
	)

	// listener
	wgListen.Add(1)
	isListening.L.Lock()
	go Listen(&wgListen, isListening, "", "12349", newConnections, shutdown)
	isListening.Wait()

	connReader, err := net.Dial(TransportProtocol, ":12349")
	if err != nil {
		assert.Fail(t, "failed to start test receiver (Dial) ", err.Error())
		return
	}

	connSender := <-newConnections

	close(shutdown)
	wgListen.Wait()

	reader := newNode(connReader, 1, output)
	reader.Wg.Add(1)
	go reader.start(done)

	sender := newNode(connSender, 1, output)
	sender.Wg.Add(1)
	go sender.start(done)

	var (
		message         = []byte{2, 1, 2, 3, 4, 5} // slot set to node slot sender
		expectedMessage = []byte{1, 1, 2, 3, 4, 5} // slot set to node slot receiver
	)

	sender.Input <- message

	timeout := time.Tick(5 * time.Second)
	select {
	case <-timeout:
		assert.Fail(t, "test timeout")
	case received := <-output:
		assert.Equal(t, expectedMessage, received, "messages sent and received are not equal")
	}

	// TODO : test closure (on both side sender & receiver) -> reception of done message
	sender.stop()
	defer func() {
		sender.Wg.Wait()
		reader.Wg.Wait()
	}()

	timeout = time.Tick(3 * time.Second)
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
