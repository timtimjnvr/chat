package conn

import (
	"fmt"
	"github.com/pkg/errors"
	"github/timtimjnvr/chat/crdt"
	"github/timtimjnvr/chat/parsestdin"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
)

const (
	localhost               = "localhost"
	localhostDecimalPointed = "127.0.0.1"
	TransportProtocol       = "tcp"

	MaxSimultaneousConnections = 100
	MaxSimultaneousMessages    = 100
	MaxMessageSize             = 1000
)

type connection struct {
	file *os.File // needed by reader package to get a file descriptor of the socket
	conn net.Conn // used to read, write bytes on the socket
}

func newConnection(conn net.Conn) (*connection, error) {
	file, err := conn.(*net.TCPConn).File()
	if err != nil {
		return nil, err
	}
	return &connection{
		conn: conn,
		file: file,
	}, nil
}

func (c *connection) Fd() uintptr {
	return c.file.Fd()
}

func (c *connection) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

func (c *connection) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

func (c *connection) Close() error {
	// close the connection
	c.conn.Close()
	// close the duplicated file
	return c.file.Close()
}

func Listen(wg *sync.WaitGroup, isReady *sync.Cond, addr, port string, newConnections chan net.Conn, shutdown chan struct{}) {
	var (
		conn      net.Conn
		wgClosure = sync.WaitGroup{}
		err       error
	)

	defer func() {
		isReady.Signal()
		wgClosure.Wait()
		wg.Done()
	}()

	ln, err := net.Listen(TransportProtocol, fmt.Sprintf("%s:%s", addr, port))
	if err != nil {
		log.Fatal("[ERROR]", err)
	}

	wgClosure.Add(1)
	go handleClosure(&wgClosure, shutdown, ln)
	isReady.Signal()

	for {
		conn, err = ln.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			log.Fatal("[ERROR]", err)
		}

		newConnections <- conn
	}
}

func InitConnections(wg *sync.WaitGroup, myInfos crdt.Infos, newJoinChatCommands <-chan parsestdin.Command, newConnections chan net.Conn, shutdown <-chan struct{}) {
	defer func() {
		wg.Done()
	}()

	for {
		select {
		case <-shutdown:
			return
		case joinChatCommand := <-newJoinChatCommands:
			args := joinChatCommand.GetArgs()
			var (
				addr     = args[parsestdin.AddrArg]
				chatRoom = args[parsestdin.ChatRoomArg]
			)

			// check if port is an int
			pt, err := strconv.Atoi(args[parsestdin.PortArg])
			if err != nil {
				log.Println(err)
			}

			/* Open connection */
			var newConn net.Conn
			newConn, err = openConnection(addr, strconv.Itoa(pt))
			if err != nil {
				log.Println("[ERROR] ", err)
				break
			}

			// init joining process
			_, err = newConn.Write(crdt.NewOperation(crdt.JoinChatByName, chatRoom, myInfos.ToBytes()).ToBytes())
			if err != nil {
				log.Println("[ERROR] ", err)
			}

			newConnections <- newConn
		}
	}
}

func openConnection(ip string, port string) (net.Conn, error) {
	if ip == localhost || ip == localhostDecimalPointed {
		ip = ""
	}

	conn, err := net.Dial(TransportProtocol, fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func handleClosure(wg *sync.WaitGroup, shutdown chan struct{}, ln net.Listener) {
	<-shutdown
	err := ln.Close()
	if err != nil {
		log.Fatal(err)
	}

	wg.Done()
}
