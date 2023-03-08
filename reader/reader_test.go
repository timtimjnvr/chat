package reader

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

var separator = []byte("\n")

func TestRead(t *testing.T) {
	var (
		maxTestDuration = 3 * time.Second
		testData        = []string{
			"first message\n",
			"second message\n",
			"third message\n",
		}
		n        int
		err      error
		w, r     *os.File
		wgReader = sync.WaitGroup{}
		messages = make(chan []byte, MaxMessageSize)
		shutdown = make(chan struct{}, 0)
	)

	w, err = os.OpenFile("test.txt", os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		assert.Fail(t, "failed to create writer (OpenFile) ", err.Error())
		return
	}

	for _, m := range testData {
		n, err = w.Write([]byte(m))
		if err != nil {
			assert.Fail(t, "failed to write ", err.Error())
			return
		}

		if n != len(m) {
			assert.Fail(t, "failed to write all the messages (Write) ", err.Error())
			return
		}
	}

	err = w.Sync()
	if err != nil {
		assert.Fail(t, "failed to sync writer (Sync) ", err.Error())
		return
	}

	r, err = os.OpenFile("test.txt", os.O_RDONLY, os.ModePerm)
	if err != nil {
		assert.Fail(t, "failed to create reader (OpenFile) ", err.Error())
		return
	}

	wgReader.Add(1)
	go Read(&wgReader, r, messages, Separator, shutdown)

	var (
		timeout = time.Tick(maxTestDuration)
		index   = 0
	)

	defer func() {
		close(shutdown)
		wgReader.Wait()
		err = os.Remove("test.txt")
		if err != nil {
			assert.Fail(t, "failed to remove file (Remove) ", err.Error())
		}
	}()

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
