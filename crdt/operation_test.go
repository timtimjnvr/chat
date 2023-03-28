package crdt

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"log"
	"reflect"
	"testing"
	"time"
)

func TestEncodeDecodeOperation(t *testing.T) {
	var (
		uuidString = "4b8e153b-834f-4190-b5d3-aba2f35ead56"
		id, _      = uuid.Parse(uuidString)
	)

	var (
		testOperations = []*Operation{
			{
				Slot:         0,
				Typology:     AddNode,
				TargetedChat: uuidString,
				Data: &NodeInfos{
					Slot:    0,
					Port:    "8080",
					Address: "localhost",
					Name:    "James",
				},
			},
			{
				Slot:         2,
				Typology:     JoinChatByName,
				TargetedChat: "my-awesome-Chat",
				Data:         &NodeInfos{Name: "Bob"},
			},
			{
				Slot:         3,
				Typology:     AddMessage,
				TargetedChat: uuidString,
				Data: &Message{
					Id:      id,
					Sender:  "James",
					Date:    time.Now(),
					Content: "Hello my Dear friend",
				},
			},
		}
	)

	for i, op := range testOperations {
		bytes := op.ToBytes()
		decodedOp, err := DecodeOperation(bytes)
		if err != nil {
			log.Println(err)
		}

		log.Println(decodedOp)
		log.Println(op)

		assert.True(t, reflect.DeepEqual(decodedOp, op), fmt.Sprintf("test %d failed to encode/decode struct", i))
	}
}

func TestGetField(t *testing.T) {
	var testData = []struct {
		bytes          []byte
		offset         int
		expectedField  []byte
		expectedOffset int
	}{
		{
			bytes:          []byte{0, 0, 0, 4, 1, 1, 1, 1, 0},
			offset:         3,
			expectedField:  []byte{1, 1, 1, 1},
			expectedOffset: 8,
		},
		{
			bytes:          []byte{0, 0, 0, 4, 1, 1, 1, 1, 2, 1, 1, 0},
			offset:         8,
			expectedField:  []byte{1, 1},
			expectedOffset: 11,
		},
	}

	for i, d := range testData {
		offset, field := getField(d.offset, d.bytes)
		assert.Equal(t, d.expectedField, field, fmt.Sprintf("test %d failed on field", i))
		assert.Equal(t, d.expectedOffset, offset, fmt.Sprintf("test %d failed on offset", i))

	}
}
