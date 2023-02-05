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
			operation{
				typology:     AddNode,
				targetedChat: uuidString,
				data:         []byte("azertyuiopqsdfghjklmwxcvbn"),
			},
			operation{
				typology:     JoinChatByName,
				targetedChat: "my-awesome-chat",
				data:         []byte("azertyuiopqsdfghjklmwxcvbn"),
			},
			operation{
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
