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

func TestNewConnection(t *testing.T) {
	conn, _, err := helperGetConnections("12340")
	if err != nil {
		assert.Fail(t, "failed to create a connections")
	}

	_, err = newConnection(conn)
	assert.NoError(t, err, "newConnection return an unexpected error")

	conn.Close()
	_, err = newConnection(conn)
	assert.Error(t, err, "newConnection on closed connection did not return an error")
}

func TestListenAndServe(t *testing.T) {
	var (
		ip              = ""
		port            = "12341"
		wgTests         = sync.WaitGroup{}
		wg              = sync.WaitGroup{}
		shutdown        = make(chan struct{}, 0)
		lock            = sync.Mutex{}
		isListening     = sync.NewCond(&lock)
		newConnections  = make(chan net.Conn, MaxSimultaneousConnections)
		maxTestDuration = 1 * time.Second
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
		go helperConnect(&wgTests, t, ip, port)
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
				return
			}
		}
	}
}

func TestInitConnections(t *testing.T) {
	var (
		listenerInfos = crdt.NewNodeInfos("127.0.0.1", "12343", "Listener")
		joinerInfos   = crdt.NewNodeInfos("127.0.0.1", "12342", "Joiner")

		wgListen          = sync.WaitGroup{}
		wgInitConnections = sync.WaitGroup{}
		shutdown          = make(chan struct{}, 0)
		lock              = sync.Mutex{}
		isListening       = sync.NewCond(&lock)

		joinChatCommands       = make(chan parsestdin.Command, MaxSimultaneousConnections)
		newConnectionsListen   = make(chan net.Conn, MaxSimultaneousConnections)
		newConnectionsInitConn = make(chan net.Conn, MaxSimultaneousConnections)
		maxTestDuration        = 1 * time.Second
	)

	defer func() {
		close(shutdown)
		wgInitConnections.Wait()
		wgListen.Wait()
	}()

	// sender
	wgListen.Add(1)
	isListening.L.Lock()
	go Listen(&wgListen, isListening, "", listenerInfos.GetPort(), newConnectionsListen, shutdown)
	isListening.Wait()

	wgInitConnections.Add(1)
	go InitConnections(&wgInitConnections, joinerInfos, joinChatCommands, newConnectionsInitConn, shutdown)

	joinChatCommand, err := parsestdin.NewCommand(fmt.Sprintf("%s %s %s %s", "/join", listenerInfos.GetAddr(), listenerInfos.GetPort(), listenerInfos.GetName()))
	if err != nil {
		assert.Fail(t, "Failed to parse command :", err.Error())
		return
	}
	for i := 0; i < MaxSimultaneousConnections; i++ {
		joinChatCommands <- joinChatCommand
	}

	var (
		timeout             = time.Tick(maxTestDuration)
		connectionsReceived int
	)

	for {
		select {
		case <-timeout:
			assert.Fail(t, "test timeout")
			return

		case newConn := <-newConnectionsListen:
			message := make([]byte, reader.MaxMessageSize)
			expectedMessage := crdt.NewOperation(crdt.JoinChatByName, "Listener", joinerInfos.ToBytes()).ToBytes()

			err = newConn.SetDeadline(time.Now().Add(maxTestDuration))
			if err != nil {
				assert.Fail(t, "Failed to set deadline on connection : ", err.Error())
			}

			var n int
			n, err = newConn.Read(message)
			if err != nil {
				assert.Fail(t, "Failed to read the connection : ", err.Error())
			}

			assert.Equal(t, expectedMessage, message[:n])
			connectionsReceived++
			if connectionsReceived == MaxSimultaneousConnections {
				return
			}
		}
	}
}

func TestReadConn(t *testing.T) {
	var (
		maxTestDuration = 1 * time.Second
		wgReader        = sync.WaitGroup{}
		messages        = make(chan []byte, reader.MaxMessageSize)
		shutdown        = make(chan struct{}, 0)
		testData        = []string{
			"first message\n",
			"second message\n",
			"third message\n",
		}
	)

	connReader, connSender, err := helperGetConnections("12345")

	for _, d := range testData {
		_, err = connSender.Write([]byte(d))
		if err != nil {
			assert.Fail(t, "failed to send bytes sender (Write) : ", err.Error())
			return
		}
	}

	var file, _ = connReader.(*net.TCPConn).File()
	wgReader.Add(1)
	go reader.Read(&wgReader, file, messages, reader.Separator, shutdown)

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

// test helper used to retrieve two linked TCP net.Conn
func helperGetConnections(port string) (net.Conn, net.Conn, error) {
	var (
		wgListen       = sync.WaitGroup{}
		shutdown       = make(chan struct{}, 0)
		lock           = sync.Mutex{}
		isListening    = sync.NewCond(&lock)
		newConnections = make(chan net.Conn, MaxSimultaneousConnections)
	)

	wgListen.Add(1)
	isListening.L.Lock()
	go Listen(&wgListen, isListening, "", port, newConnections, shutdown)
	isListening.Wait()

	conn1, err := net.Dial(TransportProtocol, fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, nil, err
	}

	conn2 := <-newConnections

	close(shutdown)
	wgListen.Wait()

	return conn1, conn2, nil
}

func helperConnect(wg *sync.WaitGroup, t *testing.T, ip, port string) {
	defer wg.Done()
	_, err := net.Dial(TransportProtocol, fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		assert.Fail(t, "failed to connect to listener : ", err.Error())
		return
	}
}
