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

	msgCommandType              commandType = iota
	connectCommandType          commandType = iota
	closeCommandType            commandType = iota
	switchDiscussionCommandType commandType = iota
	listDiscussionCommandType   commandType = iota
)

var commandTypeByCode = map[string]commandType{
	msg:              msgCommandType,
	connect:          connectCommandType,
	closeConnection:  closeCommandType,
	switchConnection: switchDiscussionCommandType,
	listDiscussion:   listDiscussionCommandType,
}

func parseCommand(text string) (command, error) {
	split := strings.Split(text, " ")
	if len(split) < 2 {
		return command{}, errors.New(errorParseCommand)
	}

	commandString := split[0]
	typology, exists := commandTypeByCode[commandString]
	if !exists {
		return command{}, errors.New(unknownCommand)
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
			return args, nil
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
