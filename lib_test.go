package main

import (
	"fmt"
	"io"
	"os"
	"syscall"
	"testing"
)

func TestTwoUsers(t *testing.T) {
	n, err := syscall.Dup(syscall.Stdout)
	if err != nil {
		t.Fatal("syscall.Dup2 err : failed to duplicate stdout", err)
		return
	}

	fmt.Println(n)
	fmt.Println("ptr", uintptr(n))

	fakeStdout := os.NewFile(uintptr(n), "fake_stdout")
	if err != nil {
		t.Fatal("os.Stdout.Close err : failed to close stdout", err)
		return
	}

	b, err := io.ReadAll(fakeStdout)
	if err != nil {
		t.Fatal("io.ReadAll err : failed to read on fake stdout", err)
		return
	}

	fmt.Println("read", string(b))
	n++
	fmt.Println(n)
	return

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

	go start("", "8080", "user1", stdinUser1)
	go start("", "8081", "user2", stdinUser2)

	_, err = w2.Write([]byte("/join localhost 8080 user1\n"))
	if err != nil {
		t.Fatal("w2.Write err : failed to write on user2 stdin", err)
		return
	}

}
