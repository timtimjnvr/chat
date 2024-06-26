package parsestdin

import (
	"fmt"
	"github/timtimjnvr/chat/crdt"
	"strings"

	"github.com/pkg/errors"
)

type (
	Command struct {
		typology crdt.OperationType
		args     map[string]string
	}
)

const (
	newChatCommand       = "/chat"
	msgCommand           = "/msg"
	joinChatCommand      = "/join"
	switchCommand        = "/switch"
	leaveChatCommand     = "/close"
	listChatUsersCommand = "/list"
	listChatsCommand     = "/list_chats"
	listAllUsersCommand  = "/list_users"
	quitCommand          = "/quit"

	MessageArg  = "messageArgument"
	PortArg     = "portArgument"
	AddrArg     = "addrArgument"
	ChatRoomArg = "chatRoomArgument"

	switchErrorSyntax  = "Command syntax :" + switchCommand + " <chat_name>"
	joinErrorSyntax    = "Command syntax : " + joinChatCommand + " <ip> <port>"
	newChatErrorSyntax = "Command syntax : " + newChatCommand + " <chat_name>"
)

var (
	commandToOperation = map[string]crdt.OperationType{
		newChatCommand:       crdt.CreateChat,
		msgCommand:           crdt.AddMessage,
		joinChatCommand:      crdt.JoinChatByName,
		switchCommand:        crdt.SwitchChat,
		leaveChatCommand:     crdt.RemoveChat,
		listChatUsersCommand: crdt.ListChatUsers,
		listAllUsersCommand:  crdt.ListUsers,
		listChatsCommand:     crdt.ListChats,
		quitCommand:          crdt.Quit,
	}

	/* PACKAGE ERRORS */

	ErrorUnknownCommand = errors.New("unknown Command")
	ErrorInArguments    = errors.New("problem in arguments")
)

func NewCommand(line string) (Command, error) {
	typology, err := parseCommandType(line)
	if err != nil {
		return Command{}, err
	}

	args, err := parseArgs(line, typology)
	if err != nil {
		return Command{}, err
	}

	return Command{
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

	case crdt.SwitchChat:
		if len(splitArgs) < 2 {
			return args, errors.Wrap(ErrorInArguments, switchErrorSyntax)
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
		messageWithoutCommand := strings.Replace(text, fmt.Sprintf("%s ", msgCommand), "", 1)
		args[MessageArg] = fmt.Sprintf("%s\n", messageWithoutCommand)

	default:
		// no args
	}

	return args, nil
}

func (c Command) GetTypology() crdt.OperationType {
	return c.typology
}

func (c Command) GetArgs() map[string]string {
	return c.args
}
