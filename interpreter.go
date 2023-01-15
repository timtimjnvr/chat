package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type (
	command struct {
		typology commandType
		args     map[string]string
	}
	commandType int
)

const (
	// COMMANDS
	msg              = "/msg"
	connect          = "/connect"
	closeConnection  = "/close"
	switchConnection = "/switch"
	listDiscussion   = "/list"

	// COMMAND TYPES
	connectCommandType          commandType = iota
	msgCommandType              commandType = iota
	closeCommandType            commandType = iota
	switchDiscussionCommandType commandType = iota
	listDiscussionCommandType   commandType = iota

	// COMMANDS ARGS
	messageArg = "messageArgument"
	portArg    = "portArgument"
	addrArg    = "addrArgument"
	idChatArg  = "idChatArgument"

	connectErrorArgumentsMsg = "command syntax : " + connect + "<ip> <port>"
)

var (
	commands = map[string]commandType{
		msg:              msgCommandType,
		connect:          connectCommandType,
		closeConnection:  closeCommandType,
		switchConnection: switchDiscussionCommandType,
		listDiscussion:   listDiscussionCommandType,
	}

	errorParseCommand   = errors.New("can't parse command")
	errorUnknownCommand = errors.New("unknown command")
	errorInArguments    = errors.New("problem in arguments")
)

func parseCommand(line string) (command, error) {
	text := fmt.Sprintf(strings.Replace(line, "\n", "", 1))
	split := strings.Split(text, " ")

	commandString := split[0]
	typology, exist := commands[commandString]

	if !exist {
		return command{}, errorUnknownCommand
	}

	args, err := getArgs(text, typology)
	if err != nil {
		return command{}, err
	}

	return command{
		typology,
		args,
	}, nil
}

func getArgs(text string, command commandType) (map[string]string, error) {
	args := make(map[string]string)
	switch command {
	case connectCommandType:
		splitArgs := strings.Split(text, " ")
		if len(splitArgs) < 3 {
			return args, errors.Wrap(errorInArguments, connectErrorArgumentsMsg)
		}

		args[addrArg] = splitArgs[1]
		args[portArg] = fmt.Sprintf(strings.Replace(splitArgs[2], "\n", "", 1))

	case msgCommandType:
		content := fmt.Sprintf(strings.Replace(text, fmt.Sprintf("%s ", msg), "", 1))

		args[messageArg] = content

	case switchDiscussionCommandType:
		content := fmt.Sprintf(strings.Replace(text, fmt.Sprintf("%s ", switchConnection), "", 1))

		_, err := strconv.Atoi(content)
		if err != nil {
			return args, errors.Wrap(err, errorInArguments.Error())
		}
		args[idChatArg] = content

	default:
		// no args
	}

	return args, nil
}
