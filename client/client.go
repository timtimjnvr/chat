package main

import (
	"client/data"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

const (
	transportProtocol      = "tcp"
	localhost              = "localhost"
	localhostDottedDecimal = "127.0.0.1"

	messageArg = "messageArgument"
	addrArg    = "portArgument"
	portArg    = "addrArgument"

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

		portAccept        = *portPtr
		addressAccept     = *addrPtr
		readerStdin       = os.Stdin
		chatList          data.Chat
		stdin             = make(chan string, maxMessagesStdin)
		newConnections    = make(chan net.Conn, maxSimultaneousConnections)
		connectionToClose = make(chan int, maxSimultaneousConnections)
	)

	defer func() {
		wgListen.Wait()
		wgReadStdin.Wait()
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
	go readStdin(&wgReadStdin, readerStdin, stdin, shutdown)

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

		case <-connectionToClose:
			// TODO

		case line := <-stdin:
			cmd, err := parseCommand(line)
			if err != nil {
				log.Println("[ERROR] ", err)
			}

			err = execute(cmd, currentChat, &chatList, newConnections, connectionToClose)

			if err != nil {
				log.Println("[ERROR] ", err)
			}
		}
	}
}

func ListenAndServe(wg *sync.WaitGroup, newConnections chan net.Conn, shutdown chan struct{}, transportProtocol, addr, port string) {
	var (
		wgConnection = sync.WaitGroup{}
		wgClosure    = sync.WaitGroup{}
		conn         net.Conn
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

		conn, err = ln.Accept()
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

func readStdin(wg *sync.WaitGroup, r *os.File, lines chan string, shutdown chan struct{}) {
	defer func() {
		wg.Done()
		close(lines)
	}()

	go func() {
		select {
		case <-shutdown:
			r.Close()
			log.Println("closed ", r.Name())
		}
	}()
	for {

		fi, err := r.Stat()
		if err != nil {
			panic(err)
		}
		if fi.Size() > 0 {
			fmt.Println("there is something to read")
			buffer := make([]byte, messageMaxSize)
			log.Println("reading")
			n, _ := r.Read(buffer)
			log.Println("finished reading")
			if n > 0 {
				lines <- string(buffer[0:n])
			}
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

func handleClosure(wg *sync.WaitGroup, shutdown chan struct{}, ln net.Listener) {
	<-shutdown
	err := ln.Close()
	if err != nil {
		log.Fatal(err)
	}

	wg.Done()
	log.Println("[INFO] handleClosure stopped")
}

func execute(cmd command, currentChat *data.Chat, chatList *data.Chat, newConnections chan net.Conn, connectionToClose chan int) error {
	var err error

	switch cmd.typology {
	case connectCommandType:
		var (
			pt     int
			addrIp string
			conn   net.Conn
		)

		pt, err = strconv.Atoi(cmd.args[portArg])
		if err != nil {
			return err
		}
		addrIp, _ = cmd.args[addrArg]
		if addrIp == localhost || addrIp == localhostDottedDecimal {
			addrIp = ""
		}

		conn, err = openConnection(transportProtocol, addrIp, pt)
		if err != nil {
			return err
		}

		newConnections <- conn

	case msgCommandType:
		err = sendMessage(currentChat.Data.Conn, cmd.args[messageArg])
		if err != nil {
			return err
		}

	case closeCommandType:
		// TODO
	case switchDiscussionCommandType:
		// TODO
	case listDiscussionCommandType:
		// TODO
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

func sendMessage(conn net.Conn, content string) error {
	buffer := []byte(content)
	_, err := conn.Write(buffer)
	if err != nil {
		return err
	}
	return nil
}
