package crdt

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEncodeDecodeMessage(t *testing.T) {
	var (
		uuidString = "4b8e153b-834f-4190-b5d3-aba2f35ead56"
		id, _      = uuid.Parse(uuidString)
	)

	var (
		tests = []struct {
			message       Message
			expectedError error
		}{
			{
				message: message{
					Id:      id,
					Sender:  "James",
					Content: "Hello my friend!",
					Date:    time.Now().UTC(),
				},
				expectedError: nil,
			},
		}
	)

	for i, test := range tests {
		bytes := test.message.ToBytes()
		res, err := DecodeMessage(bytes)
		assert.True(t, reflect.DeepEqual(res, test.message), fmt.Sprintf("test %d failed to encode/decode struct", i))
		assert.True(t, reflect.DeepEqual(err, test.expectedError), fmt.Sprintf("test %d failed on error returned", i))
	}
}
