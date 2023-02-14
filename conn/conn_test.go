package conn

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestListenAndServe(t *testing.T) {
	var (
		ip              = ""
		port            = "12345"
		wg              = sync.WaitGroup{}
		shutdown        = make(chan struct{}, 0)
		lock            = sync.Mutex{}
		isListening     = sync.NewCond(&lock)
		newConnections  = make(chan net.Conn, MaxSimultaneousConnections)
		maxTestDuration = 3 * time.Second
	)

	wg.Add(1)
	isListening.L.Lock()
	go ListenAndServe(&wg, isListening, ip, port, newConnections, shutdown)
	isListening.Wait()

	var wgTests = sync.WaitGroup{}
	for i := 0; i < MaxSimultaneousConnections; i++ {
		wgTests.Add(1)
		go connect(&wgTests, t, ip, port)
	}
	wgTests.Wait()

	var (
		timeout             = time.Tick(maxTestDuration)
		connectionsReceived = 0
	)

	select {
	case <-timeout:
		assert.Fail(t, "test timeout")
	case <-newConnections:
		connectionsReceived++
		if connectionsReceived == MaxSimultaneousConnections {
			assert.True(t, len(newConnections) == MaxSimultaneousConnections, "failed to create all connections")
		}
	}

	close(shutdown)
	wg.Wait()
}

func TestReadConn(t *testing.T) {
	var (
		maxTestDuration = 3 * time.Second
		wgReader        = sync.WaitGroup{}
		wgSender        = sync.WaitGroup{}
		messages        = make(chan []byte, MaxMessageSize)
		shutdown        = make(chan struct{}, 0)
	)

	var testData = []string{
		"first message\n",
		"second message\n",
		"third message\n",
	}

	// sender
	wgSender.Add(1)
	go func(wgSender *sync.WaitGroup, shutdown chan struct{}) {
		defer wgSender.Done()
		ln, err := net.Listen(transportProtocol, ":12348")
		if err != nil {
			assert.Fail(t, "failed to start test sender (Listen) : ", err.Error())
		}

		connSender, err := ln.Accept()
		if err != nil {
			assert.Fail(t, "failed to start test sender (Accept) : ", err.Error())
		}

		for _, d := range testData {
			connSender.Write([]byte(d))
		}

		select {
		case <-shutdown:
			connSender.Close()
			return
		}

	}(&wgSender, shutdown)

	connReader, err := net.Dial(transportProtocol, ":12348")
	if err != nil {
		assert.Fail(t, "failed to start test receiver (Dial) : ", err.Error())
	}

	wgReader.Add(1)
	go read(&wgReader, connReader, messages, shutdown)

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

	close(shutdown)
	wgSender.Wait()
	wgReader.Wait()
}

func connect(wg *sync.WaitGroup, t *testing.T, ip, port string) {
	defer wg.Done()
	_, err := net.Dial(transportProtocol, fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		assert.Fail(t, "failed to connect to listener")
		return
	}
}
