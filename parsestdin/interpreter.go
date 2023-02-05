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
	Command interface {
		getCommandType() crdt.OperationType
		getCommandArgs() map[string]string
	}

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

	connectErrorArgumentsMsg = "command syntax : " + joinChatCommand + "<ip> <port>"

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

func newCommand(line string) (Command, error) {
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

func (c command) getCommandType() crdt.OperationType {
	return c.typology
}

func (c command) getCommandArgs() map[string]string {
	return c.args
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

func parseArgs(text string, command crdt.OperationType) (map[string]string, error) {
	args := make(map[string]string)

	switch command {
	case crdt.JoinChatByName:
		splitArgs := strings.Split(text, " ")
		if len(splitArgs) < 3 {
			return args, errors.Wrap(ErrorInArguments, connectErrorArgumentsMsg)
		}
		args[AddrArg] = removeSubStrings(splitArgs[1], " ", "\n")
		args[PortArg] = removeSubStrings(splitArgs[2], " ", "\n")

	case crdt.AddMessage:
		args[MessageArg] = removeSubStrings(text, fmt.Sprintf("%s ", msgCommand), "\n")

	default:
		// no args
	}

	return args, nil
}

func removeSubStrings(source string, patterns ...string) string {
	var result = source
	for _, pattern := range patterns {
		result = strings.Replace(result, pattern, "", -1)
	}

	return result
}

func ReadStdin(wg *sync.WaitGroup, stdin chan<- []byte, shutdown chan struct{}) {
	defer func() {
		close(stdin)
		wg.Done()
		log.Println("[INFO] readStdin stopped")
	}()

	// writeClose is closed in order to signal readStdin stop signal
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
			stdin <- buffer[0:n]
		}
	}
}

func HandleStdin(wg *sync.WaitGroup, myInfos crdt.Infos, newConnections chan<- net.Conn, newOperations chan<- []byte, shutdown chan struct{}) {
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
	go ReadStdin(&wgReadStdin, stdin, shutdown)

	for {
		select {
		case <-shutdown:
			return

		case line := <-stdin:
			cmd, err := newCommand(string(line))
			if err != nil {
				log.Println("[ERROR] ", err)
			}

			args := cmd.getCommandArgs()

			switch typology := cmd.getCommandType(); typology {
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
				newOperations <- newChatOperationBytes

			case crdt.JoinChatByName:
				var (
					addr     = args[AddrArg]
					chatRoom = args[ChatRoomArg]
					pt       int
				)

				pt, err = strconv.Atoi(args[PortArg])
				if err != nil {
					log.Println(err)
				}

				/* Open connection */
				var newConn net.Conn
				newConn, err = conn.OpenConnection(addr, pt)
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

				newConnections <- newConn

			case crdt.AddMessage:
				content := args[MessageArg]
				if currentChat == nil {
					log.Println(noDiscussionSelected)
					continue
				}

				/* Add the message to discussion & sync with other nodes */
				var message []byte
				message, err = crdt.NewMessage(myInfos.GetName(), content).ToBytes()
				if err != nil {
					log.Println("[ERROR] ", err)
				}

				syncMessage := crdt.NewOperation(crdt.AddMessage, currentChat.GetId(), message)
				bytesSyncMessage := []byte(string(syncMessage.ToBytes()))

				newOperations <- bytesSyncMessage

			case crdt.LeaveChat:
				/* TODO
				leave discussion and gracefully shutdown connection with all nodes
				*/

			case crdt.Quit:
				/* TODO
				leave all discussions and gracefully shutdown connection with all nodes
				*/
			}

		}
	}

}
