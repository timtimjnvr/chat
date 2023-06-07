package reader

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"syscall"
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
		readerDone = make(chan struct{})
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

	go Read(readerDone, r, messages, Separator, shutdown)

	var (
		timeout = time.Tick(maxTestDuration)
		index   = 0
	)

	defer func() {
		close(shutdown)
		err = os.Remove("test.txt")
		if err != nil {
			assert.Fail(t, "failed to remove file (Remove) ", err.Error())
		}
	}()

	for {
		if index == len(testData) {
			break
		}

		select {
		case <-timeout:
			assert.Fail(t, "test timeout shutting done with shutdown")
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
	_ = r.Close()

	select {
	case <-timeout:
		assert.Fail(t, "test timeout waiting reader")
		return

	case <-readerDone:
		// pass
		return
	}
}

func TestRead_SOMAXCONN(t *testing.T) {
	var (
		maxTestDuration = 3 * time.Second
		w, r     *os.File
		shutdown      = make(chan struct{}, 0)
		testsDoneChan = make(map[int]chan struct{})
		err           error
	)

	for i := 0; i < syscall.SOMAXCONN; i++ {
		file := fmt.Sprintf("test_%d.txt", i)
		w, err = os.OpenFile(file, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			assert.Fail(t, "failed to create writer (OpenFile) ", err.Error())
			return
		}

		var (
			messages = make(chan []byte, MaxMessageSize)
			m        = []byte("test message\n")
			n        int
		)

		n, err = w.Write(m)
		if err != nil {
			assert.Fail(t, "failed to write ", err.Error())
			return
		}

		if n != len(m) {
			assert.Fail(t, "failed to write the message (Write) ", err.Error())
			return
		}

		err = w.Sync()
		if err != nil {
			assert.Fail(t, "failed to sync writer (Sync) ", err.Error())
			return
		}

		var done = make(chan struct{})

		r, err = os.OpenFile(file, os.O_RDONLY, os.ModePerm)
		if err != nil {
			assert.Fail(t, "failed to create reader (OpenFile) ", err.Error())
			return
		}

		go Read(done, r, messages, Separator, shutdown)
		testsDoneChan[i] = done
	}

	close(shutdown)
	timeout := time.Tick(maxTestDuration)
	for i := 0; i < syscall.SOMAXCONN; i++ {
		select {
			case <-timeout:
				assert.Fail(t, "test timeout waiting for reader")
			case <-testsDoneChan[i]:

				file := fmt.Sprintf("test_%d.txt", i)
				err = os.Remove(file)
				if err != nil {
					assert.Fail(t, "failed to remove file (Remove) ", err.Error())
				}
		}
	}
}
