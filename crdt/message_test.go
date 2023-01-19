package crdt

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetSyncBytes(t *testing.T) {

	var (
		id, _ = uuid.Parse("4b8e153b-834f-4190-b5d3-aba2f35ead56")

		m = &message{
			id:      id,
			Sender:  "toto",
			Content: "J'aime",
			Date:    time.Time{},
		}

		tests = []struct {
			inputStruct   Operable
			operation     operationType
			expectedRunes []rune
		}{
			{
				inputStruct:   m,
				operation:     add,
				expectedRunes: []rune{0, 0, 36, 52, 98, 56, 101, 49, 53, 51, 98, 45, 56, 51, 52, 102, 45, 52, 49, 57, 48, 45, 98, 53, 100, 51, 45, 97, 98, 97, 50, 102, 51, 53, 101, 97, 100, 53, 54, 4, 116, 111, 116, 111, 6, 74, 39, 97, 105, 109, 101, 20, 48, 48, 48, 49, 45, 48, 49, 45, 48, 49, 84, 48, 48, 58, 48, 48, 58, 48, 48, 90},
			},
		}
	)

	for i, test := range tests {
		res := test.inputStruct.toRunes(test.operation)
		assert.True(t, reflect.DeepEqual(res, test.expectedRunes), fmt.Sprintf("test %d failed on runes returned", i))
	}
}

func TestGetMessageFromBytes(t *testing.T) {

	var (
		id, _ = uuid.Parse("4b8e153b-834f-4190-b5d3-aba2f35ead56")

		m = message{
			id:      id,
			Sender:  "toto",
			Content: "J'aime",
			Date:    time.Time{},
		}

		tests = []struct {
			runes          []rune
			expectedStruct interface{}
		}{
			{
				runes:          []rune{36, 52, 98, 56, 101, 49, 53, 51, 98, 45, 56, 51, 52, 102, 45, 52, 49, 57, 48, 45, 98, 53, 100, 51, 45, 97, 98, 97, 50, 102, 51, 53, 101, 97, 100, 53, 54, 4, 116, 111, 116, 111, 6, 74, 39, 97, 105, 109, 101, 20, 48, 48, 48, 49, 45, 48, 49, 45, 48, 49, 84, 48, 48, 58, 48, 48, 58, 48, 48, 90},
				expectedStruct: m,
			},
		}
	)
	for i, test := range tests {
		res := GetMessageFromBytes(test.runes)
		assert.True(t, reflect.DeepEqual(res, test.expectedStruct), fmt.Sprintf("test %d failed on struct returned", i))
	}
}
