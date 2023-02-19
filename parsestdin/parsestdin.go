package parsestdin

import (
	"github/timtimjnvr/chat/conn"
	"github/timtimjnvr/chat/crdt"
	"github/timtimjnvr/chat/reader"

	"fmt"
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

	logFrmt     = "[INFO] %s\n"
	typeCommand = "type a command :"
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

func HandleStdin(wg *sync.WaitGroup, file *os.File, myInfos crdt.Infos, connCreated chan<- net.Conn, operationsCreated chan<- crdt.Operation, shutdown chan struct{}) {
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
	go reader.ReadFile(&wgReadStdin, file, stdin, shutdown)

	fmt.Printf(logFrmt, typeCommand)

	for {
		select {
		case <-shutdown:
			return

		case line := <-stdin:
			fmt.Printf(logFrmt, typeCommand)

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

				operationsCreated <- crdt.NewOperation(crdt.CreateChat, newChat.GetId(), bytesChat)

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

				err = conn.Send(newConn, crdt.NewOperation(crdt.JoinChatByName, chatRoom, myInfos.ToBytes()).ToBytes())
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

				/* Add the messageBytes to discussion & sync with other nodes */
				var messageBytes []byte
				messageBytes = crdt.NewMessage(myInfos.GetName(), content).ToBytes()
				operationsCreated <- crdt.NewOperation(crdt.AddMessage, currentChat.GetId(), messageBytes)

			case crdt.LeaveChat:
				operationsCreated <- crdt.NewOperation(crdt.LeaveChat, currentChat.GetId(), myInfos.ToBytes())

			case crdt.Quit:
				operationsCreated <- crdt.NewOperation(crdt.Quit, "", myInfos.ToBytes())
			}
		}
	}
}
