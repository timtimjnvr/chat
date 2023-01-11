package main

import (
	"chat/data"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

const (
	ip                      = "ip"
	port                    = "port"
	transportProtocol       = "tcp"
	localhost               = "localhost"
	localhostDecimalPointed = "127.0.0.1"

	messageArg = "messageArgument"
	portArg    = "portArgument"
	addrArg    = "portArgument"

	endOfStream                = "Ctrl+D\n\r"
	maxSimultaneousConnections = 1000
	messageMaxSize             = 10000
	maxMessagesStdin           = 100
	maxMessagesConn            = 100
)

func main() {
	portPtr := flag.String("p", "8080", "port number used to accept connection")
	addrPtr := flag.String("a", "", "address used to accept connection")
	flag.Parse()

	var (
		currentChat *data.Chat
		wgListen    = sync.WaitGroup{}
		wgReadStdin = sync.WaitGroup{}

		sigc     = make(chan os.Signal, 1)
		shutdown = make(chan struct{})

		portAccept    = *portPtr
		addressAccept = *addrPtr

		chatList       data.Chat
		stdin          = make(chan string, maxMessagesStdin)
		newConnections = make(chan net.Conn, maxSimultaneousConnections)
	)

	defer func() {
		wgReadStdin.Wait()
		wgListen.Wait()
		log.Println("[INFO] program shutdown")
	}()

	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	wgListen.Add(1)
	go ListenAndServe(&wgListen, newConnections, shutdown, transportProtocol, addressAccept, portAccept)

	wgReadStdin.Add(1)
	go readStdin(&wgReadStdin, stdin, shutdown)

	for {
		select {
		case <-sigc:
			close(shutdown)
			return

		case conn := <-newConnections:
			// add to chat list
			wg := sync.WaitGroup{}
			chat := data.AddChat(&chatList, &wg, conn)
			currentChat = chat
			go handleConnection(chat, shutdown)

		case line := <-stdin:
			cmd, err := parseCommand(line)
			if err != nil {
				log.Println("[ERROR] ", err)
			}

			switch cmd.typology {
			case connectCommandType:
				var (
					pt, _ = strconv.Atoi(cmd.args[portArg])
					addr  = cmd.args[portArg]
					conn  net.Conn
				)

				if addr == localhost || addr == localhostDecimalPointed {
					addr = ""
				}

				conn, err = openConnection(transportProtocol, addr, pt)
				if err != nil {
					log.Println("[ERROR] ", err)
				}

				newConnections <- conn

			case msgCommandType:
				content := cmd.args[messageArg]
				err = sendMessage(currentChat, content)
				if err != nil {
					log.Println("[ERROR] ", err)
				}

			case closeCommandType:
				// TODO

			case switchDiscussionCommandType:
				//TODO
			}
		}
	}
}

func ListenAndServe(wg *sync.WaitGroup, newConnections chan net.Conn, shutdown chan struct{}, transportProtocol, addr, port string) {
	var (
		wgConnection = sync.WaitGroup{}
		wgClosure    = sync.WaitGroup{}
	)

	defer func() {
		wgClosure.Wait()
		wgConnection.Wait()
		wg.Done()
		log.Println("[INFO] ListenAndServe stopped")
	}()

	ln, err := net.Listen(transportProtocol, fmt.Sprintf("%s:%s", addr, port))
	if err != nil {
		log.Fatal(err)
	}

	wgClosure.Add(1)
	go handleClosure(&wgClosure, shutdown, ln)

	for {
		log.Println(fmt.Sprintf("[INFO] Accepting connections on port %s", port))

		conn, err := ln.Accept()
		if err != nil && errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			log.Println("[WARNING] err Accept :", err)
			continue
		}

		wgConnection.Add(1)
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
			buffer = make([]byte, messageMaxSize)
			err    error
		)

		fdSet.Clear(int(os.Stdin.Fd()))
		fdSet.Clear(int(readClose.Fd()))

		fdSet.Set(int(os.Stdin.Fd()))
		fdSet.Set(int(readClose.Fd()))

		// modifies r/w/e file descriptors in fdSet with ready to use file descriptors (ie for us stdin or readClose)
		_, err = unix.Select(int(readClose.Fd()+1), &fdSet, nil, nil, &unix.Timeval{Sec: 60, Usec: 0})
		if err != nil {
			log.Fatal("[ERROR] ", err)
			return
		}

		// shutdown
		if fdSet.IsSet(int(readClose.Fd())) {
			return
		}

		// default read stdin
		var n int
		n, err = os.Stdin.Read(buffer)
		if err != nil {
			log.Fatal("[ERROR] ", err)
			return
		}
		if n > 0 {
			lines <- string(buffer[0:n])
		}
	}
}

// return value should be turned into a <- chan msg
func readConn(wg *sync.WaitGroup, conn net.Conn, shutdown chan struct{}) <-chan string {
	defer wg.Done()

	messages := make(chan string, maxMessagesConn)
	go func() {
		defer close(messages)
		for {
			select {
			case <-shutdown:
				log.Println("[INFO] readConn shutting down")
				return
			default:
				buffer := make([]byte, messageMaxSize)
				n, _ := conn.Read(buffer)
				if n > 0 {
					messages <- string(buffer[0:n])
				}
			}
		}
	}()
	return messages
}

func handleConnection(chat *data.Chat, shutdown chan struct{}) error {
	log.Println("[INFO] new connection")

	var (
		wg   = sync.WaitGroup{}
		conn = chat.Data.Conn
	)

	defer func() {
		wg.Wait()
		chat.Data.Wg.Done()
		log.Println("[INFO] handleConnection stopped")
	}()

	wg.Add(1)
	messages := readConn(&wg, conn, shutdown)

	for {
		select {
		case <-shutdown:
			return nil
		case message := <-messages:
			log.Println("[INFO]: received msg", message)
		default:
			// pass
		}
	}

	return nil
}

func openConnection(protocol, ip string, port int) (net.Conn, error) {
	conn, err := net.Dial(protocol, fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func sendMessage(chat *data.Chat, content string) error {
	buffer := []byte(content)
	_, err := chat.Data.Conn.Write(buffer)
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
