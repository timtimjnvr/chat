package parsestdin

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github/timtimjnvr/chat/crdt"
	"reflect"
	"testing"
)

func TestParseCommandType(t *testing.T) {
	t.Parallel()

	ass := assert.New(t)

	var tests = []struct {
		line             string
		expectedTypology crdt.OperationType
		expectedErr      error
	}{
		{
			line:             "/chat ***********\n\n",
			expectedTypology: crdt.CreateChat,
			expectedErr:      nil,
		},
		{
			line:             "/msg ***********!\n\n",
			expectedTypology: crdt.AddMessage,
			expectedErr:      nil,
		},
		{
			line:             "/join ***********!\n",
			expectedTypology: crdt.JoinChatByName,
			expectedErr:      nil,
		},
		{
			line:             "/close ***********!\n\n",
			expectedTypology: crdt.LeaveChat,
			expectedErr:      nil,
		},
		{
			line:             "/list ***********!\n\n",
			expectedTypology: crdt.ListUsers,
			expectedErr:      nil,
		},
		{
			line:             "/quit ***********!\n\n",
			expectedTypology: crdt.Quit,
			expectedErr:      nil,
		},
		{
			line:             "/quit**********\n",
			expectedTypology: *new(crdt.OperationType),
			expectedErr:      ErrorUnknownCommand,
		},
		{
			line:             "/unknown **********\n",
			expectedTypology: *new(crdt.OperationType),
			expectedErr:      ErrorUnknownCommand,
		},
	}

	for i, test := range tests {
		typology, err := parseCommandType(test.line)
		ass.Equal(typology, test.expectedTypology, fmt.Sprintf("test %d failed on computing typology", i))
		ass.True(errors.Is(err, test.expectedErr), fmt.Sprintf("test %d failed on error returned", i))
	}

}

func TestGetArgs(t *testing.T) {
	t.Parallel()

	ass := assert.New(t)

	var tests = []struct {
		text         string
		typology     crdt.OperationType
		expectedArgs map[string]string
		expectedErr  error
	}{
		{
			text:         "/chat my-awesome-chat\n",
			typology:     crdt.CreateChat,
			expectedArgs: map[string]string{ChatRoomArg: "my-awesome-chat"},
			expectedErr:  nil,
		},
		{
			text:         "/msg Hello friend!\n",
			typology:     crdt.AddMessage,
			expectedArgs: map[string]string{MessageArg: "Hello friend!"},
			expectedErr:  nil,
		},
		{
			text:         "/join 127.0.0.1 8080 my-awesome-chat\n",
			typology:     crdt.JoinChatByName,
			expectedArgs: map[string]string{AddrArg: "127.0.0.1", PortArg: "8080", ChatRoomArg: "my-awesome-chat"},
			expectedErr:  nil,
		},
		{
			text:         "/join 127.0.0.1\n",
			typology:     crdt.JoinChatByName,
			expectedArgs: make(map[string]string),
			expectedErr:  ErrorInArguments,
		},
		{
			text:         "/list\n",
			typology:     crdt.ListUsers,
			expectedArgs: make(map[string]string),
			expectedErr:  nil,
		},
		{
			text:         "/close\n",
			typology:     crdt.LeaveChat,
			expectedArgs: make(map[string]string),
			expectedErr:  nil,
		},
		{
			text:         "/quit\n",
			typology:     crdt.Quit,
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
