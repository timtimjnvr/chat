package parsestdin

import (
	"fmt"
	"strconv"
	"strings"

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

	CreateChatCommandType       commandType = iota
	ConnectCommandType          commandType = iota
	MsgCommandType              commandType = iota
	CloseCommandType            commandType = iota
	SwitchDiscussionCommandType commandType = iota
	ListDiscussionCommandType   commandType = iota
	QuitCommandType             commandType = iota

	/* COMMANDS ARGS */

	MessageArg = "messageArgument"
	PortArg    = "portArgument"
	AddrArg    = "addrArgument"
	IdChatArg  = "idChatArgument"

	/* INLINE COMMANDS */

	chat             = "/chat"
	msg              = "/msg"
	connect          = "/connect"
	closeConnection  = "/close"
	switchConnection = "/switch"
	listDiscussion   = "/list"
	quit             = "/quit"

	/* ERRORS FORMAT */

	connectErrorArgumentsMsg = "command syntax : " + connect + "<ip> <port>"
)

var (
	commands = map[string]commandType{
		msg:              MsgCommandType,
		connect:          ConnectCommandType,
		closeConnection:  CloseCommandType,
		switchConnection: SwitchDiscussionCommandType,
		listDiscussion:   ListDiscussionCommandType,
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

		args[AddrArg] = splitArgs[1]
		args[PortArg] = fmt.Sprintf(strings.Replace(splitArgs[2], "\n", "", 1))

	case MsgCommandType:
		content := fmt.Sprintf(strings.Replace(text, fmt.Sprintf("%s ", msg), "", 1))

		args[MessageArg] = content

	case SwitchDiscussionCommandType:
		content := fmt.Sprintf(strings.Replace(text, fmt.Sprintf("%s ", switchConnection), "", 1))

		_, err := strconv.Atoi(content)
		if err != nil {
			return args, errors.Wrap(err, ErrorInArguments.Error())
		}

		args[IdChatArg] = content

	default:
		// no args
	}

	return args, nil
}
