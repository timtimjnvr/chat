package parsestdin

import (
	"chat/crdt"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"log"
	"reflect"
	"testing"
)

func TestParseCommandType(t *testing.T) {
	ass := assert.New(t)

	var tests = []struct {
		line             string
		expectedTypology crdt.OperationType
		expectedErr      error
	}{
		{
			line:             "/msg blabla\n",
			expectedTypology: crdt.AddMessage,
			expectedErr:      nil,
		},
		{
			line:             "/toto blabla\n",
			expectedTypology: *new(crdt.OperationType),
			expectedErr:      ErrorUnknownCommand,
		},
	}

	for i, test := range tests {
		typology, err := parseCommandType(test.line)
		log.Println(typology, test.expectedTypology)
		ass.Equal(typology, test.expectedTypology, fmt.Sprintf("test %d failed on computing typology", i))
		ass.True(errors.Is(err, test.expectedErr), fmt.Sprintf("test %d failed on error returned", i))
	}

}

func TestGetArgs(t *testing.T) {
	ass := assert.New(t)

	var tests = []struct {
		text         string
		typology     crdt.OperationType
		expectedArgs map[string]string
		expectedErr  error
	}{
		{
			text:         "/msg content blabla\n",
			typology:     crdt.AddMessage,
			expectedArgs: map[string]string{MessageArg: "content blabla"},
			expectedErr:  nil,
		},
		{
			text:         "/join 127.0.0.1 8080\n",
			typology:     crdt.JoinChatByName,
			expectedArgs: map[string]string{AddrArg: "127.0.0.1", PortArg: "8080"},
			expectedErr:  nil,
		},
		{
			text:         "/join 127.0.0.1\n",
			typology:     crdt.JoinChatByName,
			expectedArgs: make(map[string]string),
			expectedErr:  ErrorInArguments,
		},
		{
			text:         "/close\n",
			typology:     crdt.LeaveChat,
			expectedArgs: make(map[string]string),
			expectedErr:  nil,
		},
		{
			text:         "/list\n",
			typology:     crdt.ListUsers,
			expectedArgs: make(map[string]string),
			expectedErr:  nil,
		},
	}

	for i, test := range tests {
		args, err := parseArgs(test.text, test.typology)
		ass.True(reflect.DeepEqual(args, test.expectedArgs), fmt.Sprintf("test %d failed on computing arguments", i))
		ass.True(errors.Is(err, test.expectedErr), fmt.Sprintf("test %d failed on error returned", i))
	}
}
