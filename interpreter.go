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

	connectCommandType          commandType = iota
	msgCommandType              commandType = iota
	closeCommandType            commandType = iota
	switchDiscussionCommandType commandType = iota
	listDiscussionCommandType   commandType = iota

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
	errorUnknownCommand = errors.New("unknow command")
	errorInArguments    = errors.New("problem in arguments")
)

func parseCommand(text string) (command, error) {
	split := strings.Split(text, " ")
	if len(split) < 2 {
		return command{}, errorParseCommand
	}

	commandString := split[0]
	typology, exist := commands[commandString]
	if !exist {
		return command{}, errorUnknownCommand
	}

	args, err := getArgs(text, typology)
	if err != nil {
		return command{}, errorInArguments
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
		if len(splitArgs) <= 3 {
			return args, errors.Wrap(errorInArguments, connectErrorArgumentsMsg)
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

	return args, nil
}
