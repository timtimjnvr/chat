package parsestdin

import (
	"chat/conn"
	"fmt"
	"golang.org/x/sys/unix"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

type (
	Command interface {
		GetCommandType() commandType
		GetCommandArgs() map[string]string
	}

	command struct {
		typology commandType
		args     map[string]string
	}

	commandType int
)

func (c command) GetCommandType() commandType {
	return c.typology
}

func (c command) GetCommandArgs() map[string]string {
	return c.args
}

func NewCommand(line string) (Command, error) {
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

const (
	/* COMMAND TYPES*/

	CreateChatCommandType commandType = iota
	ConnectCommandType    commandType = iota
	MsgCommandType        commandType = iota
	CloseCommandType      commandType = iota
	ListUsersCommandType  commandType = iota
	QuitCommandType       commandType = iota

	/* COMMANDS ARGS */

	MessageArg  = "messageArgument"
	PortArg     = "portArgument"
	AddrArg     = "addrArgument"
	IdChatArg   = "idChatArgument"
	ChatRoomArg = "chatRoomArgument"

	/* INLINE COMMANDS */

	chat            = "/chat"
	msg             = "/msg"
	connect         = "/connect"
	closeConnection = "/close"
	listUsers       = "/list"
	quit            = "/quit"

	/* ERRORS FORMAT */

	connectErrorArgumentsMsg = "command syntax : " + connect + "<ip> <port>"
)

var (
	commands = map[string]commandType{
		chat:            CreateChatCommandType,
		msg:             MsgCommandType,
		connect:         ConnectCommandType,
		closeConnection: CloseCommandType,
		listUsers:       ListUsersCommandType,
		quit:            QuitCommandType,
	}

	/* PACKAGE ERRORS */

	ErrorParseCommand   = errors.New("can't parse command")
	ErrorUnknownCommand = errors.New("unknown command")
	ErrorInArguments    = errors.New("problem in arguments")
)

func parseCommandType(line string) (commandType, error) {
	text := fmt.Sprintf(strings.Replace(line, "\n", "", 1))
	split := strings.Split(text, " ")

	commandString := split[0]
	typology, exist := commands[commandString]

	if !exist {
		return typology, ErrorUnknownCommand
	}

	return typology, nil
}

func parseArgs(text string, command commandType) (map[string]string, error) {
	args := make(map[string]string)

	switch command {
	case ConnectCommandType:
		splitArgs := strings.Split(text, " ")
		if len(splitArgs) < 3 {
			return args, errors.Wrap(ErrorInArguments, connectErrorArgumentsMsg)
		}
		args[AddrArg] = removeSubStrings(splitArgs[1], " ", "\n")
		args[PortArg] = removeSubStrings(splitArgs[2], " ", "\n")

	case MsgCommandType:
		args[MessageArg] = removeSubStrings(text, fmt.Sprintf("%s ", msg), "\n")

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
			writeClose.Close()
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
