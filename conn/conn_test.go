package conn

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"sync"
	"testing"
)

func TestListenAndServe(t *testing.T) {
	var (
		numberOfTest   = 10
		ip             = ""
		port           = "8090"
		wg             = sync.WaitGroup{}
		shutdown       = make(chan struct{}, 0)
		lock           = sync.Mutex{}
		isListening    = sync.NewCond(&lock)
		newConnections = make(chan net.Conn, numberOfTest)
		err            error
	)

	wg.Add(1)

	isListening.L.Lock()
	go ListenAndServe(&wg, isListening, ip, port, newConnections, shutdown)
	isListening.Wait()

	var wgTests = sync.WaitGroup{}
	for i := 0; i < numberOfTest; i++ {
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

	log.Println(len(newConnections))

	assert.True(t, len(newConnections) == numberOfTest, "failed to create all connections")
	close(shutdown)
	wg.Wait()
}
