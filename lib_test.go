package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"math/rand"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestTwoUsers(t *testing.T) {
	_, err := os.OpenFile("debug.txt", os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		assert.Fail(t, "failed to create writer (OpenFile) ", err.Error())
		return
	}

	// keep backup of the real stdout
	oldStdout := os.Stdout
	rStdout, wStdout, _ := os.Pipe()
	os.Stdout = wStdout

	// get stdout content in the background
	var (
		stop = make(chan struct{})
		outC = make(chan string, 1000)
		wg   = &sync.WaitGroup{}
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return

			default:
				var buf bytes.Buffer
				_, _ = io.Copy(&buf, rStdout)
				outC <- buf.String()
			}
		}
	}()

	// TODO define list of expected display messages

	// Set up test resources and start test scenarios
	rand.Seed(time.Now().UnixNano())
	var (
		port1 = rand.Intn(65533-49152) + 49152
		port2 = port1 + 1
		sigC1 = make(chan os.Signal, 1)
		sigC2 = make(chan os.Signal, 1)
	)

	stdinUser1, _, err := os.Pipe()
	if err != nil {
		t.Fatal("os.Pipe err : failed to set up user 1", err)
		return
	}

	stdinUser2, w2, err := os.Pipe()
	if err != nil {
		t.Fatal("os.Pipe err : failed to set up user 2", err)
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		start("", fmt.Sprintf("%d", port1), "user1", stdinUser1, sigC1)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		start("", fmt.Sprintf("%d", port2), "user2", stdinUser2, sigC2)
	}()

	// TODO send the commands to simulated clients
	_, err = w2.Write([]byte(fmt.Sprintf("/join localhost %d user1\n", port1)))
	if err != nil {
		t.Fatal("w2.Write err : failed to write on user2 stdin", err)
		return
	}

	// stop simulation
	sigC1 <- syscall.SIGINT
	sigC2 <- syscall.SIGINT

	// stop reading redirected stdout
	// wait for simulation to be done
	close(stop)
	wg.Wait()

	// restoring stdout for printing test results
	wStdout.Close()
	os.Stdout = oldStdout
	go func() {
		for m := range outC {
			// TODO compare with expected
			fmt.Println(m)
		}
	}()
}
