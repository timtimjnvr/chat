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
		port           = "12345"
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
			log.Println("connecting")
			_, err = net.Dial(transportProtocol, fmt.Sprintf("%s:%s", ip, port))
			if err != nil {
				log.Println("ERR", err)
				assert.Fail(t, "failed to connect to listener")
				return
			}

		}(&wgTests)
	}

	wgTests.Wait()
	close(shutdown)
	wg.Wait()

	assert.True(t, len(newConnections) == numberOfTest, "failed to create all connections")
	log.Println(len(newConnections))
	log.Println("TEST END")
}
