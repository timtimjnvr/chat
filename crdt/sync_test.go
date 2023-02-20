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
		testOperations = []Operation{
			&operation{
				slot:         0,
				typology:     AddNode,
				targetedChat: uuidString,
				data:         []byte("azertyuiopqsdfghjklmwxcvbn"),
			},
			&operation{
				slot:         2,
				typology:     JoinChatByName,
				targetedChat: "my-awesome-chat",
				data:         []byte("azertyuiopqsdfghjklmwxcvbn"),
			},
			&operation{
				slot:         3,
				typology:     AddMessage,
				targetedChat: uuidString,
				data: message{
					Id:      id,
					Sender:  "James",
					Date:    time.Now(),
					Content: "Hello my Dear friend",
				}.ToBytes(),
			},
		}
	)

	for i, op := range testOperations {
		bytes := op.ToBytes()
		res := DecodeOperation(bytes)
		assert.True(t, reflect.DeepEqual(res, op), fmt.Sprintf("test %d failed to encode/decode struct", i))
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
