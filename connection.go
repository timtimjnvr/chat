package main

import (
	"chat/node"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
	"log"
	"net"
	"os"
	"sync"
)

func listenAndServe(wg *sync.WaitGroup, newConnections chan net.Conn, shutdown chan struct{}, transportProtocol, addr, port string) {
	var (
		conn      net.Conn
		wgClosure = sync.WaitGroup{}
		err       error
	)

	defer func() {
		wgClosure.Wait()
		wg.Done()
		log.Println("[INFO] listenAndServe stopped")
	}()

	ln, err := net.Listen(transportProtocol, fmt.Sprintf("%s:%s", addr, port))
	if err != nil {
		log.Fatal(err)
	}

	wgClosure.Add(1)
	go handleClosure(&wgClosure, shutdown, ln)
	for {
		log.Println(fmt.Sprintf("[INFO] Accepting connections on port %s", port))
		conn, err = ln.Accept()
		if err != nil && errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			log.Println("[WARNING] err Accept :", err)
			continue
		}

		newConnections <- conn
	}
}

func readStdin(wg *sync.WaitGroup, lines chan string, shutdown chan struct{}) {
	defer func() {
		close(lines)
		wg.Done()
		log.Println("[INFO] readStdin stopped")
	}()

	// writeClose is closed in order to signal readStdin stop signal
	var readClose, writeClose, _ = os.Pipe()

	go func() {
		select {
		case <-shutdown:
			writeClose.Close()
		}
	}()

	for {
		log.Println("[INFO] type a command")

		var (
			fdSet  = unix.FdSet{}
			buffer = make([]byte, maxMessageSize)
			err    error
		)

		fdSet.Clear(int(os.Stdin.Fd()))
		fdSet.Clear(int(readClose.Fd()))

		fdSet.Set(int(os.Stdin.Fd()))
		fdSet.Set(int(readClose.Fd()))

		// modifies r/w/e file descriptors in fdSet with ready to use file descriptors (ie for us parsestdin or readClose)
		_, err = unix.Select(int(readClose.Fd()+1), &fdSet, nil, nil, &unix.Timeval{Sec: 60, Usec: 0})
		if err != nil {
			log.Fatal("[ERROR] ", err)
			return
		}

		// shutdown
		if fdSet.IsSet(int(readClose.Fd())) {
			return
		}

		// default read parsestdin
		var n int
		n, err = os.Stdin.Read(buffer)
		if err != nil {
			return
		}

		if n > 0 {
			lines <- string(buffer[0:n])
		}
	}
}

func readConn(wg *sync.WaitGroup, conn net.Conn, messages chan string, shutdown chan struct{}) {
	defer func() {
		close(messages)
		wg.Done()
		log.Println("[INFO] readConn stopped")
	}()

	for {
		select {
		case <-shutdown:
			return

		default:
			buffer := make([]byte, maxMessageSize)
			n, err := conn.Read(buffer)
			if err != nil {
				return
			}
			if n > 0 {
				log.Print(string(buffer))
				messages <- string(buffer[0:n])
			}
		}
	}
}

func handleConnection(wg *sync.WaitGroup, conn net.Conn, id uuid.UUID, done chan<- uuid.UUID, shutdown chan struct{}) {
	log.Println("[INFO] new connection")

	var (
		wgReadConn = sync.WaitGroup{}
		messages   = make(chan string)
	)

	defer func() {
		conn.Close()
		wgReadConn.Wait()
		wg.Done()
		done <- id
		log.Println("[INFO] connection lost for ", conn.LocalAddr())
	}()

	wgReadConn.Add(1)
	go readConn(&wgReadConn, conn, messages, shutdown)

	for {
		select {
		case <-shutdown:
			return

		case message, ok := <-messages:
			if !ok {
				// connection closed on the other side
				return
			}
			log.Println("[INFO]: received message : ", message)
		}
	}
}

func openConnection(protocol, ip string, port int) (net.Conn, error) {
	conn, err := net.Dial(protocol, fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func sendMessage(chat *node.Node, content string) error {
	buffer := []byte(content)
	_, err := chat.Business.Conn.Write(buffer)
	if err != nil {
		return err
	}
	return nil
}

func handleClosure(wg *sync.WaitGroup, shutdown chan struct{}, ln net.Listener) {
	<-shutdown
	err := ln.Close()
	if err != nil {
		log.Fatal(err)
	}

	wg.Done()
	log.Println("[INFO] handleClosure stopped")
}
