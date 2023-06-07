package conn

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github/timtimjnvr/chat/crdt"
	"github/timtimjnvr/chat/reader"
	"net"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestListenAndServe(t *testing.T) {
	var (
		ip              = ""
		port            = "12341"
		wgTests         = sync.WaitGroup{}
		wg              = sync.WaitGroup{}
		shutdown        = make(chan struct{}, 0)
		lock            = sync.Mutex{}
		isListening     = sync.NewCond(&lock)
		newConnections  = make(chan net.Conn)
		maxTestDuration = 1 * time.Second
	)

	defer func() {
		close(shutdown)
		wgTests.Wait()
		wg.Wait()
	}()

	wg.Add(1)
	isListening.L.Lock()
	go CreateConnections(&wg, isListening, &crdt.NodeInfos{Address: ip, Port: port}, make(chan ConnectionRequest), newConnections, shutdown)
	isListening.Wait()

	for i := 0; i < syscall.SOMAXCONN; i++ {
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

		case c := <-newConnections:
			connectionsReceived++
			c.Close()
			if connectionsReceived == syscall.SOMAXCONN {
				return
			}
		}
	}
}

func TestConnect(t *testing.T) {
	var (
		listenerInfos = crdt.NewNodeInfos("127.0.0.1", "12343", "Listener")
		joinerInfos   = crdt.NewNodeInfos("127.0.0.1", "12342", "Joiner")

		wgListen    = sync.WaitGroup{}
		wgConnect   = sync.WaitGroup{}
		shutdown    = make(chan struct{}, 0)
		lock        = sync.Mutex{}
		isListening = sync.NewCond(&lock)

		connectionRequests     = make(chan ConnectionRequest)
		newConnectionsListen   = make(chan net.Conn)
		newConnectionsInitConn = make(chan net.Conn)
		maxTestDuration        = 1 * time.Second
	)

	defer func() {
		close(shutdown)
		wgListen.Wait()
		wgConnect.Wait()
	}()

	wgListen.Add(1)
	isListening.L.Lock()
	go CreateConnections(&wgListen, isListening, &crdt.NodeInfos{Address: "", Port: listenerInfos.Port}, make(chan ConnectionRequest), newConnectionsListen, shutdown)
	isListening.Wait()

	wgConnect.Add(1)
	go Connect(&wgConnect, joinerInfos, connectionRequests, newConnectionsInitConn, shutdown)

	connRequest := NewConnectionRequest(listenerInfos.Port, listenerInfos.Address, listenerInfos.Name)

	for i := 0; i < syscall.SOMAXCONN; i++ {
		connectionRequests <- connRequest
		c := <-newConnectionsInitConn
		c.Close()
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

		case c := <-newConnectionsListen:
			message := make([]byte, reader.MaxMessageSize)
			expectedMessage := crdt.NewOperation(crdt.JoinChatByName, "Listener", joinerInfos).ToBytes()

			err := c.SetDeadline(time.Now().Add(maxTestDuration))
			if err != nil {
				assert.Fail(t, "Failed to set deadline on conn : ", err.Error())
			}

			var n int
			n, err = c.Read(message)
			if err != nil {
				assert.Fail(t, "Failed to read the conn : ", err.Error())
			}

			assert.Equal(t, expectedMessage, message[:n])
			connectionsReceived++

			c.Close()
			if connectionsReceived == syscall.SOMAXCONN {
				return
			}
		}
	}
}

func TestReadConn(t *testing.T) {
	var (
		maxTestDuration = 1 * time.Second
		wgReader   = sync.WaitGroup{}
		readerDone = make(chan struct{})
		messages   = make(chan []byte)
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

	c, err := newConn(connReader)
	if err != nil {
		assert.Fail(t, "failed to create connection reader : ", err.Error())
		return
	}

	go reader.Read(readerDone, c, messages, reader.Separator, shutdown)

	defer func() {
		close(shutdown)
		wgReader.Wait()
	}()

	var (
		timeout = time.Tick(maxTestDuration)
		index   = 0
	)

	for {
		if index == len(testData) {
			break
		}
		select {
		case <-timeout:
			assert.Fail(t, "test timeout")
			return

		case m := <-messages:
			assert.Equal(t, strings.TrimSuffix(testData[index], "\n"), string(m), "message differs")
			index++
			if index == len(testData) {
				continue
			}
		}
	}

	timeout = time.Tick(maxTestDuration)
	connSender.Close()

	select {
	case <-timeout:
		assert.Fail(t, "test timeout waiting reader")
		return

	case <-readerDone:
		// pass
	}

}

// test helper used to retrieve two linked TCP net.Conn
func helperGetConnections(port string) (net.Conn, net.Conn, error) {
	var (
		wgListen       = sync.WaitGroup{}
		shutdown       = make(chan struct{}, 0)
		lock           = sync.Mutex{}
		isListening    = sync.NewCond(&lock)
		newConnections = make(chan net.Conn)
	)

	wgListen.Add(1)
	isListening.L.Lock()
	go CreateConnections(&wgListen, isListening, &crdt.NodeInfos{Address: "", Port: port}, make(chan ConnectionRequest), newConnections, shutdown)
	isListening.Wait()

	conn1, err := net.Dial(transportProtocol, fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, nil, err
	}

	conn2 := <-newConnections

	close(shutdown)
	wgListen.Wait()

	return conn1, conn2, nil
}

func helperConnect(wg *sync.WaitGroup, t *testing.T, ip, port string) {
	var c net.Conn

	defer func() {
		wg.Done()
		c.Close()
	}()

	c, err := net.Dial(transportProtocol, fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		assert.Fail(t, "failed to connect to listener : ", err.Error())
		return
	}
}
