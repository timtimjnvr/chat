package crdt

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
		tests = []struct {
			op          *Operation
			expectedErr error
		}{
			{
				&Operation{
					Slot:         0,
					Typology:     AddChat,
					TargetedChat: "my-awesome-chat",
					Data: &Chat{
						Id:   "9b83358e-a570-4a1b-8842-6800ee770f2a",
						Name: "james",
						Messages: []*Message{
							{
								Id:      uuid.New(),
								Sender:  "james",
								Content: "1\n",
								Date:    "2023-04-30T23:53:33+02:00",
							},
						},
					},
				},
				nil,
			},
			{
				&Operation{
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
				nil,
			},
			{
				&Operation{
					Slot:         2,
					Typology:     JoinChatByName,
					TargetedChat: "my-awesome-Chat",
					Data: &NodeInfos{
						Slot:    0,
						Port:    "8080",
						Address: "localhost",
						Name:    "James",
					},
				},
				nil,
			},
			{
				&Operation{
					Slot:         3,
					Typology:     AddMessage,
					TargetedChat: uuidString,
					Data: &Message{
						Id:      id,
						Sender:  "James",
						Date:    time.Now().Format(time.RFC3339),
						Content: "Hello my Dear friend",
					},
				},
				nil,
			},
		}
	)

	for i, test := range tests {
		bytes := test.op.ToBytes()
		decodedOp, err := DecodeOperation(bytes)
		if err != nil {
			assert.Equal(t, err, test.expectedErr, fmt.Sprintf("test %d failed on getting error", i))
		}

		assert.True(t, reflect.DeepEqual(decodedOp, test.op), fmt.Sprintf("test %d failed to encode/decode struct", i))
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
