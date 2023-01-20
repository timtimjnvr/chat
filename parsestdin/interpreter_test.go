package parsestdin

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestParseCommandType(t *testing.T) {
	ass := assert.New(t)

	var tests = []struct {
		line             string
		expectedTypology commandType
		expectedErr      error
	}{
		{
			line:             "/msg blabla\n",
			expectedTypology: MsgCommandType,
			expectedErr:      nil,
		},
		{
			line:             "/toto blabla\n",
			expectedTypology: *new(commandType),
			expectedErr:      ErrorUnknownCommand,
		},
	}

	for i, test := range tests {
		typology, err := parseCommandType(test.line)
		ass.True(reflect.DeepEqual(typology, test.expectedTypology), fmt.Sprintf("test %d failed on computing typology", i))
		ass.True(errors.Is(err, test.expectedErr), fmt.Sprintf("test %d failed on error returned", i))
	}

}

func TestGetArgs(t *testing.T) {
	ass := assert.New(t)

	var tests = []struct {
		text         string
		typology     commandType
		expectedArgs map[string]string
		expectedErr  error
	}{
		{
			text:         "/msg content blabla\n",
			typology:     MsgCommandType,
			expectedArgs: map[string]string{MessageArg: "content blabla"},
			expectedErr:  nil,
		},
		{
			text:         "/connect 127.0.0.1 8080\n",
			typology:     ConnectCommandType,
			expectedArgs: map[string]string{AddrArg: "127.0.0.1", PortArg: "8080"},
			expectedErr:  nil,
		},
		{
			text:         "/connect 127.0.0.1\n",
			typology:     ConnectCommandType,
			expectedArgs: make(map[string]string),
			expectedErr:  ErrorInArguments,
		},
		{
			text:         "/close\n",
			typology:     CloseCommandType,
			expectedArgs: make(map[string]string),
			expectedErr:  nil,
		},
		{
			text:         "/list\n",
			typology:     ListUsersCommandType,
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
