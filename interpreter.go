package main

import (
	"fmt"
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
	msg              = "/msg"
	connect          = "/connect"
	closeConnection  = "/close"
	switchConnection = "/switch"
	listDiscussion   = "/list"

	errorParseCommand = "can't parse command"
	unknownCommand    = "unknow command"

	connectCommandType          commandType = iota
	msgCommandType              commandType = iota
	closeCommandType            commandType = iota
	switchDiscussionCommandType commandType = iota
	listDiscussionCommandType   commandType = iota
)

func parseCommand(text string) (command, error) {
	split := strings.Split(text, " ")
	if len(split) < 2 {
		return command{}, errors.New(errorParseCommand)
	}

	commandString := split[0]
	switch commandString {
	case connect:
		args := getArgs(text, connectCommandType)

		return command{
			connectCommandType,
			args,
		}, nil

	case msg:
		args := getArgs(text, connectCommandType)

		return command{
			msgCommandType,
			args,
		}, nil

	case switchConnection:
		args := getArgs(text, switchDiscussionCommandType)

		return command{
			switchDiscussionCommandType,
			args,
		}, nil

	case closeConnection:
		args := getArgs(text, closeCommandType)

		return command{
			closeCommandType,
			args,
		}, nil

	case listDiscussion:
		return command{
			listDiscussionCommandType,
			nil,
		}, nil

	default:
		return command{}, errors.New(unknownCommand)
	}
}

func getArgs(text string, command commandType) map[string]string {
	args := make(map[string]string)
	switch command {
	case connectCommandType:
		splitArgs := strings.Split(text, " ")
		if len(splitArgs) <= 3 {
			return args
		}

		args[addrArg] = splitArgs[1]
		args[portArg] = splitArgs[2]

	case msgCommandType:
		content := fmt.Sprintf(strings.Replace(text, fmt.Sprintf("%s ", msg), "", 1))
		args[messageArg] = content

	case closeCommandType:
		// TODO

	case switchDiscussionCommandType:
		//TODO
	}

	return args
}
