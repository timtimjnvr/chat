package parsestdin

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func testParseCommandType(t *testing.T) {
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

func testGetArgs(t *testing.T) {
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
			expectedArgs: map[string]string{MessageArg: "content content blabla"},
			expectedErr:  nil,
		},
		{
			text:         "/connect 127.0.0.1 8080\n",
			typology:     MsgCommandType,
			expectedArgs: map[string]string{AddrArg: " 127.0.0.1", PortArg: "8080"},
			expectedErr:  nil,
		},
		{
			text:         "/connect 127.0.0.1\n",
			typology:     MsgCommandType,
			expectedArgs: map[string]string{},
			expectedErr:  ErrorInArguments,
		},
		{
			text:         "/close\n",
			typology:     CloseCommandType,
			expectedArgs: map[string]string{},
			expectedErr:  nil,
		},
		{
			text:         "/list\n",
			typology:     ListDiscussionCommandType,
			expectedArgs: map[string]string{},
			expectedErr:  nil,
		},
		{
			text:         "/switch 3\n",
			typology:     ListDiscussionCommandType,
			expectedArgs: map[string]string{IdChatArg: "3"},
			expectedErr:  nil,
		},
		{
			text:         "/switch 3 \n",
			typology:     ListDiscussionCommandType,
			expectedArgs: map[string]string{IdChatArg: "3"},
			expectedErr:  nil,
		},
		{
			text:         "/switch aa \n",
			typology:     ListDiscussionCommandType,
			expectedArgs: map[string]string{},
			expectedErr:  ErrorInArguments,
		},
	}

	for i, test := range tests {
		args, err := parseArgs(test.text, test.typology)
		ass.True(reflect.DeepEqual(args, test.expectedArgs), fmt.Sprintf("test %d failed on computing arguments", i))
		ass.True(errors.Is(err, test.expectedErr), fmt.Sprintf("test %d failed on error returned", i))
	}
}
