package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"math/rand"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestTwoUsers(t *testing.T) {
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

	go func() {
		defer close(outC)
		for {
			select {
			case <-stop:
				return

			default:
				var buf = bytes.Buffer{}
				_, err := io.Copy(&buf, rStdout)
				if err != nil {
					return
				}
				outC <- buf.String()
			}
		}
	}()

	var expectedMessages = []string{
		"[INFO] type a Command :",
		"[INFO] type a Command :",
		"[INFO] type a Command :",
		"[INFO] user2 joined chat",
		"[INFO] you joined a new chat : user1",
		"[INFO] type a Command :",
		"user2 (2024-05-19T16:45:02+02:00): hey man",
		"user2 (2024-05-19T16:45:02+02:00): hey man",
		"[INFO] program shutdown",
		"[INFO] program shutdown",
	}

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

		start("", fmt.Sprintf("%d", port1), "user1", stdinUser1, sigC1, false)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		start("", fmt.Sprintf("%d", port2), "user2", stdinUser2, sigC2, false)
	}()

	_, err = w2.Write([]byte(fmt.Sprintf("/join localhost %d user1\n", port1)))
	if err != nil {
		t.Fatal("w2.Write err : failed to write on user2 stdin", err)
		return
	}

	<-time.Tick(1 * time.Second)

	_, err = w2.Write([]byte("/msg hey man"))
	if err != nil {
		t.Fatal("w2.Write err : failed to write on user2 stdin", err)
		return
	}

	<-time.Tick(5 * time.Second)

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

	wg.Add(1)
	go func() {
		defer wg.Done()
		index := 0
		for m := range outC {
			messages := strings.Split(m, "\n")
			assert.Equal(t, expectedMessages[index], messages[index])
		}
	}()
	wg.Wait()
}
