package crdt

import (
	"fmt"
	"log"
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
			Id:      id,
			Sender:  "toto",
			Content: "J'aime",
			Date:    time.Time{},
		}

		tests = []struct {
			message       Message
			operation     OperationType
			expectedbytes []byte
		}{
			{
				message:       m,
				operation:     AddMessage,
				expectedbytes: []byte{123, 34, 105, 100, 34, 58, 34, 52, 98, 56, 101, 49, 53, 51, 98, 45, 56, 51, 52, 102, 45, 52, 49, 57, 48, 45, 98, 53, 100, 51, 45, 97, 98, 97, 50, 102, 51, 53, 101, 97, 100, 53, 54, 34, 44, 34, 115, 101, 110, 100, 101, 114, 34, 58, 34, 116, 111, 116, 111, 34, 44, 34, 99, 111, 110, 116, 101, 110, 116, 34, 58, 34, 74, 39, 97, 105, 109, 101, 34, 44, 34, 100, 97, 116, 101, 34, 58, 34, 48, 48, 48, 49, 45, 48, 49, 45, 48, 49, 84, 48, 48, 58, 48, 48, 58, 48, 48, 90, 34, 125},
			},
		}
	)

	for i, test := range tests {
		res, _ := test.message.ToBytes()
		log.Println(res)
		assert.True(t, reflect.DeepEqual(res, test.expectedbytes), fmt.Sprintf("test %d failed on runes returned", i))
	}
}

func TestGetMessageFromBytes(t *testing.T) {

	var (
		id, _ = uuid.Parse("4b8e153b-834f-4190-b5d3-aba2f35ead56")

		m = message{
			Id:      id,
			Sender:  "toto",
			Content: "J'aime",
			Date:    time.Time{},
		}

		tests = []struct {
			bytes           []byte
			expectedMessage message
		}{
			{
				bytes:           []byte{123, 34, 105, 100, 34, 58, 34, 52, 98, 56, 101, 49, 53, 51, 98, 45, 56, 51, 52, 102, 45, 52, 49, 57, 48, 45, 98, 53, 100, 51, 45, 97, 98, 97, 50, 102, 51, 53, 101, 97, 100, 53, 54, 34, 44, 34, 115, 101, 110, 100, 101, 114, 34, 58, 34, 116, 111, 116, 111, 34, 44, 34, 99, 111, 110, 116, 101, 110, 116, 34, 58, 34, 74, 39, 97, 105, 109, 101, 34, 44, 34, 100, 97, 116, 101, 34, 58, 34, 48, 48, 48, 49, 45, 48, 49, 45, 48, 49, 84, 48, 48, 58, 48, 48, 58, 48, 48, 90, 34, 125},
				expectedMessage: m,
			},
		}
	)
	for i, test := range tests {
		res, _ := DecodeMessage(test.bytes)
		assert.True(t, reflect.DeepEqual(res, test.expectedMessage), fmt.Sprintf("test %d failed on struct returned", i))
	}
}
