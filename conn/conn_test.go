package conn

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net"
	"sync"
	"testing"
)

func TestListenAndServe(t *testing.T) {
	var (
		ip             = ""
		port           = "8080"
		wg             = sync.WaitGroup{}
		shutdown       = make(chan struct{}, 0)
		lock           = sync.Mutex{}
		isListening    = sync.NewCond(&lock)
		newConnections = make(chan net.Conn, MaxSimultaneousConnections)
		err            error
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
			}
		}(&wgTests)
	}
	wgTests.Wait()

	assert.True(t, len(newConnections) == MaxSimultaneousConnections, "failed to create all connections")
	close(shutdown)
	wg.Wait()
}
