package parsestdin

import (
	"chat/conn"
	"chat/crdt"

	"fmt"
	"golang.org/x/sys/unix"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

type (
	command struct {
		typology crdt.OperationType
		args     map[string]string
	}
)

const (
	newChatCommand   = "/chat"
	msgCommand       = "/msg"
	joinChatCommand  = "/join"
	leaveChatCommand = "/close"
	listUsersCommand = "/list"
	quitCommand      = "/quit"

	MessageArg  = "messageArgument"
	PortArg     = "portArgument"
	AddrArg     = "addrArgument"
	ChatRoomArg = "chatRoomArgument"

	joinErrorSyntax    = "command syntax : " + joinChatCommand + " <ip> <port>"
	newChatErrorSyntax = "command syntax : " + newChatCommand + " <chat_name>"

	maxMessagesStdin     = 100
	noDiscussionSelected = "you must be in a discussion to send a message"
)

var (
	commandToOperation = map[string]crdt.OperationType{
		newChatCommand:   crdt.CreateChat,
		msgCommand:       crdt.AddMessage,
		joinChatCommand:  crdt.JoinChatByName,
		leaveChatCommand: crdt.LeaveChat,
		listUsersCommand: crdt.ListUsers,
		quitCommand:      crdt.Quit,
	}

	/* PACKAGE ERRORS */

	ErrorUnknownCommand = errors.New("unknown command")
	ErrorInArguments    = errors.New("problem in arguments")
)

func newCommand(line string) (command, error) {
	typology, err := parseCommandType(line)
	if err != nil {
		return command{}, err
	}

	args, err := parseArgs(line, typology)
	if err != nil {
		return command{}, err
	}

	return command{
		typology,
		args,
	}, nil
}

func parseCommandType(line string) (crdt.OperationType, error) {
	text := fmt.Sprintf(strings.Replace(line, "\n", "", 1))
	split := strings.Split(text, " ")

	commandString := split[0]
	operationTypology, exist := commandToOperation[commandString]

	if !exist {
		return operationTypology, ErrorUnknownCommand
	}

	return operationTypology, nil
}

func parseArgs(line string, command crdt.OperationType) (map[string]string, error) {
	var (
		text      = fmt.Sprintf(strings.Replace(line, "\n", "", 1))
		args      = make(map[string]string)
		splitArgs = strings.Split(text, " ")
	)

	switch command {
	case crdt.CreateChat:
		// no chat room specified
		if len(splitArgs) < 2 {
			return args, errors.Wrap(ErrorInArguments, newChatErrorSyntax)
		}

		args[ChatRoomArg] = strings.Replace(splitArgs[1], " ", "", 2)

	case crdt.JoinChatByName:
		// not enough args
		if len(splitArgs) <= 3 {
			return args, errors.Wrap(ErrorInArguments, joinErrorSyntax)
		}

		args[AddrArg] = strings.Replace(splitArgs[1], " ", "", 2)
		args[PortArg] = strings.Replace(splitArgs[2], " ", "", 2)
		args[ChatRoomArg] = strings.Replace(splitArgs[3], " ", "", 2)

	case crdt.AddMessage:
		args[MessageArg] = strings.Replace(text, fmt.Sprintf("%s ", msgCommand), "", 1)

	default:
		// no args
	}

	return args, nil
}

func readStdin(wg *sync.WaitGroup, stdin chan<- []byte, shutdown chan struct{}) {
	defer func() {
		close(stdin)
		wg.Done()
		log.Println("[INFO] readStdin stopped")
	}()

	// writeClose is closed in order to signal to stop reading stdin
	var readClose, writeClose, _ = os.Pipe()

	go func() {
		select {
		case <-shutdown:
			_ = writeClose.Close()
		}
	}()

	for {
		log.Println("[INFO] type a command")

		var (
			fdSet  = unix.FdSet{}
			buffer = make([]byte, conn.MaxMessageSize)
			err    error
		)

		fdSet.Clear(int(os.Stdin.Fd()))
		fdSet.Clear(int(readClose.Fd()))

		fdSet.Set(int(os.Stdin.Fd()))
		fdSet.Set(int(readClose.Fd()))

		// wait and modifies file descriptors in fdSet with first ready to use file descriptors (ie for us stdin or readClose)
		_, err = unix.Select(int(readClose.Fd()+1), &fdSet, nil, nil, &unix.Timeval{Sec: 60, Usec: 0})
		if err != nil {
			log.Fatal("[ERROR] ", err)
			return
		}

		// readClose : stop reading stdin
		if fdSet.IsSet(int(readClose.Fd())) {
			return
		}

		// default : read stdin
		var n int
		n, err = os.Stdin.Read(buffer)
		if err != nil {
			return
		}

		if n > 0 {
			stdin <- buffer[0:n]
		}
	}
}

func HandleStdin(wg *sync.WaitGroup, myInfos crdt.Infos, connCreated chan<- net.Conn, operationsCreated chan<- []byte, shutdown chan struct{}) {
	var (
		wgReadStdin = sync.WaitGroup{}
		currentChat = crdt.NewChat(myInfos.GetName())
	)

	defer func() {
		wgReadStdin.Wait()
		wg.Done()
	}()

	wgReadStdin.Add(1)
	var stdin = make(chan []byte, maxMessagesStdin)
	go readStdin(&wgReadStdin, stdin, shutdown)

	for {
		select {
		case <-shutdown:
			return

		case line := <-stdin:
			cmd, err := newCommand(string(line))
			if err != nil {
				log.Println("[ERROR] ", err)
			}

			args := cmd.args

			switch cmd.typology {
			case crdt.CreateChat:
				var (
					bytesChat []byte
					chatName  = args[ChatRoomArg]
					newChat   = crdt.NewChat(chatName)
				)
				bytesChat, err = newChat.ToBytes()
				if err != nil {
					log.Println(err)
				}

				newChatOperation := crdt.NewOperation(crdt.CreateChat, newChat.GetId(), bytesChat)
				newChatOperationBytes := newChatOperation.ToBytes()
				operationsCreated <- newChatOperationBytes

			case crdt.JoinChatByName:
				var (
					addr     = args[AddrArg]
					chatRoom = args[ChatRoomArg]
					pt       int
				)

				// check if port is an int
				pt, err = strconv.Atoi(args[PortArg])
				if err != nil {
					log.Println(err)
				}

				/* Open connection */
				var newConn net.Conn
				newConn, err = conn.OpenConnection(addr, strconv.Itoa(pt))
				if err != nil {
					log.Println("[ERROR] ", err)
					break
				}

				myInfosBytes, _ := myInfos.ToBytes()
				joinOperation := crdt.NewOperation(crdt.JoinChatByName, chatRoom, myInfosBytes)
				err = conn.Send(newConn, joinOperation.ToBytes())
				if err != nil {
					log.Println("[ERROR] ", err)
				}

				connCreated <- newConn

			case crdt.AddMessage:
				content := args[MessageArg]
				if currentChat == nil {
					log.Println(noDiscussionSelected)
					continue
				}

				/* Add the message to discussion & sync with other nodes */
				var message []byte
				message = crdt.NewMessage(myInfos.GetName(), content).ToBytes()

				addMessageSync := crdt.NewOperation(crdt.AddMessage, currentChat.GetId(), message).ToBytes()
				operationsCreated <- addMessageSync

			case crdt.LeaveChat:
				myInfosBytes, _ := myInfos.ToBytes()
				if err != nil {
					log.Println("[ERROR] ", err)
				}
				leaveChatSync := crdt.NewOperation(crdt.LeaveChat, currentChat.GetId(), myInfosBytes).ToBytes()
				operationsCreated <- leaveChatSync

			case crdt.Quit:
				myInfosBytes, _ := myInfos.ToBytes()
				if err != nil {
					log.Println("[ERROR] ", err)
				}
				quitSync := crdt.NewOperation(crdt.Quit, "", myInfosBytes).ToBytes()
				operationsCreated <- quitSync
			}
		}
	}
}
