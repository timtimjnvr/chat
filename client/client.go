package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	portAccept  = "8080"
	portConnect = "8080"

	transportProtocol = "tcp"
)

func main() {
	var (
		wg       = sync.WaitGroup{}
		sigc     = make(chan os.Signal, 1)
		shutdown = make(chan struct{})
	)

	defer func() {
		wg.Wait()
		log.Println("[INFO] program shutdown")
	}()

	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	wg.Add(1)
	go ListenAndServe(&wg, shutdown, transportProtocol, portAccept)

	for {
		select {
		case <-sigc:
			close(shutdown)
			return
		default:
			// pass
		}
	}
}

func ListenAndServe(wg *sync.WaitGroup, shutdown chan struct{}, transportProtocol string, port string) {
	var wgClosure = sync.WaitGroup{}

	defer func() {
		wgClosure.Wait()
		wg.Done()
		log.Println("[INFO] ListenAndServe stopped")

	}()

	ln, err := net.Listen(transportProtocol, ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	wgClosure.Add(1)
	go handleClosure(&wgClosure, shutdown, ln)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) error {
	log.Println("[INFO] new connection")

	for {
		buffer := make([]byte, 100)
		n, err := conn.Read(buffer)
		if err != nil {
			return err
		}
		if n > 0 {
			log.Println(string(buffer))
			break
		}
	}

	log.Println("[INFO] handleClosure stopped")
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
