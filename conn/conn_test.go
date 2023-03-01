package conn

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github/timtimjnvr/chat/crdt"
	"github/timtimjnvr/chat/parsestdin"
	"github/timtimjnvr/chat/reader"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestListenAndServe(t *testing.T) {
	var (
		ip              = ""
		port            = "8091"
	 	wgTests = sync.WaitGroup{}
		wg              = sync.WaitGroup{}
		shutdown        = make(chan struct{}, 0)
		lock            = sync.Mutex{}
		isListening     = sync.NewCond(&lock)
		newConnections  = make(chan net.Conn, MaxSimultaneousConnections)
		maxTestDuration = 3 * time.Second
	)

	defer func() {
		close(shutdown)
		wgTests.Wait()
		wg.Wait()
	}()

	wg.Add(1)
	isListening.L.Lock()
	go Listen(&wg, isListening, ip, port, newConnections, shutdown)
	isListening.Wait()

	for i := 0; i < MaxSimultaneousConnections; i++ {
		wgTests.Add(1)
		go connect(&wgTests, t, ip, port)
	}

	var (
		timeout             = time.Tick(maxTestDuration)
		connectionsReceived = 0
	)

	for {
		select {
		case <-timeout:
			assert.Fail(t, "test timeout")
			return

		case <-newConnections:
			connectionsReceived++
			if connectionsReceived == MaxSimultaneousConnections {
				assert.True(t, connectionsReceived == MaxSimultaneousConnections, "failed to create all connections")
				return
			}
		}
	}
}

func TestReadConn(t *testing.T) {
	var (
		maxTestDuration = 3 * time.Second
		wgReader        = sync.WaitGroup{}
		messages        = make(chan []byte, reader.MaxMessageSize)
		wgListen        = sync.WaitGroup{}
		shutdown        = make(chan struct{}, 0)
		lock            = sync.Mutex{}
		isListening     = sync.NewCond(&lock)
		newConnections  = make(chan net.Conn, MaxSimultaneousConnections)
	)

	var testData = []string{
		"first message\n",
		"second message\n",
		"third message\n",
	}

	// sender
	wgListen.Add(1)
	isListening.L.Lock()
	go Listen(&wgListen, isListening, "", "12348", newConnections, shutdown)
	isListening.Wait()

	connReader, err := net.Dial(transportProtocol, ":12348")
	if err != nil {
		assert.Fail(t, "failed to start test receiver (Dial) ", err.Error())
		return
	}

	connSender := <-newConnections
	for _, d := range testData {
		_, err = connSender.Write([]byte(d))
		if err != nil {
			assert.Fail(t, "failed to send bytes sender (Write) ", err.Error())
			return
		}
	}

	var file, _ = connReader.(*net.TCPConn).File()
	wgReader.Add(1)
	go reader.ReadFile(&wgReader, file, messages, shutdown)

	defer func() {
		close(shutdown)
		wgReader.Wait()
	}()

	var (
		timeout = time.Tick(maxTestDuration)
		index   = 0
	)

	for {
		select {
		case <-timeout:
			assert.Fail(t, "test timeout")
			return

		case m := <-messages:
			assert.Equal(t, strings.TrimSuffix(testData[index], "\n"), string(m), "message differs")
			index++
			if index == len(testData) {
				return
			}
		}
	}
}

func TestInitNodeConnections(t *testing.T) {
	var (
		listenerInfos = crdt.NewNodeInfos("127.0.0.1", "12348", "Listener")
		joinerInfos   = crdt.NewNodeInfos("127.0.0.1", "12349", "Joiner")

		wgListen          = sync.WaitGroup{}
		wgInitConnections = sync.WaitGroup{}
		shutdown          = make(chan struct{}, 0)
		lock              = sync.Mutex{}
		isListening       = sync.NewCond(&lock)

		joinChatCommands  = make(chan parsestdin.Command, 1)
		newConnections    = make(chan net.Conn, 1)

		maxTestDuration = 3 * time.Second
	)

	defer func() {
		close(shutdown)
		wgInitConnections.Wait()
		wgListen.Wait()
	}()

	// sender
	wgListen.Add(1)
	isListening.L.Lock()
	go Listen(&wgListen, isListening, "", "12348", newConnections, shutdown)
	isListening.Wait()

	wgInitConnections.Add(1)
	go InitNodeConnections(&wgInitConnections, joinerInfos, joinChatCommands, newConnections, shutdown)

	joinChatCommand, err := parsestdin.NewCommand(fmt.Sprintf("%s %s %s %s", "/join", listenerInfos.GetAddr(), listenerInfos.GetPort(), listenerInfos.GetName()))
	if err != nil {
		assert.Fail(t, "Failed to parse command")
    	return
	}

	joinChatCommands <- joinChatCommand
	timeout := time.Tick(maxTestDuration)

	for {
		select {
		case <-timeout:
			assert.Fail(t, "test timeout")
			return

		case newConn := <-newConnections:
			message := make([]byte, reader.MaxMessageSize)
			expectedMessage := crdt.NewOperation(crdt.JoinChatByName, "Listener", joinerInfos.ToBytes()).ToBytes()

			err = newConn.SetDeadline(time.Now().Add(maxTestDuration))
			if err != nil {
				assert.Fail(t, "Failed to set deadline on connection")
			}

			var n int
			n, err = newConn.Read(message)
			if err != nil {
				assert.Fail(t, "Failed to read the connection")
			}

			assert.Equal(t, expectedMessage, message[:n])
			return
		}
	}
}

func connect(wg *sync.WaitGroup, t *testing.T, ip, port string) {
	defer wg.Done()
	_, err := net.Dial(transportProtocol, fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		assert.Fail(t, "failed to connect to listener")
		return
	}
}
