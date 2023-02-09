package conn

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net"
	"sync"
	"testing"
	"time"
)

func TestListenAndServe(t *testing.T) {
	var (
		maxTestDuration = 3 * time.Second
		ip              = ""
		port            = "12345"
		wg              = sync.WaitGroup{}
		shutdown        = make(chan struct{}, 0)
		lock            = sync.Mutex{}
		isListening     = sync.NewCond(&lock)
		newConnections  = make(chan net.Conn, MaxSimultaneousConnections)
		err             error
	)

	wg.Add(1)
	isListening.L.Lock()
	go ListenAndServe(&wg, isListening, ip, port, newConnections, shutdown)
	isListening.Wait()

	var wgTests = sync.WaitGroup{}
	for i := 0; i < MaxSimultaneousConnections; i++ {
		wgTests.Add(1)
		go func(wgTests *sync.WaitGroup) {
			defer wgTests.Done()
			_, err = net.Dial(transportProtocol, fmt.Sprintf("%s:%s", ip, port))
			if err != nil {
				assert.Fail(t, "failed to connect to listener")
				return
			}

		}(&wgTests)
	}
	wgTests.Wait()

	timeout := time.Tick(maxTestDuration)
	received := 0
	select {
	case <-timeout:
		assert.Fail(t, "test timeout")
	case <-newConnections:
		received++
		if received == MaxSimultaneousConnections {
			assert.True(t, len(newConnections) == MaxSimultaneousConnections, "failed to create all connections")
		}
	}

	close(shutdown)
	wg.Wait()
}
