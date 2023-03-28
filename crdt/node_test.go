package crdt

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestEncodeDecodeNode(t *testing.T) {
	var (
		uuidString = "4b8e153b-834f-4190-b5d3-aba2f35ead56"
		id, _      = uuid.Parse(uuidString)
	)

	var (
		tests = []struct {
			message       *NodeInfos
			expectedError error
		}{
			{
				message: &NodeInfos{
					Slot:    1,
					Id:      id,
					Port:    "8080",
					Address: "localhost",
					Name:    "James",
				},
				expectedError: nil,
			},
		}
	)

	for i, test := range tests {
		bytes := test.message.ToBytes()
		var res = &NodeInfos{}
		err := DecodeData(bytes, res)
		assert.True(t, reflect.DeepEqual(res, test.message), fmt.Sprintf("test %d failed to encode/decode struct", i))
		assert.True(t, reflect.DeepEqual(err, test.expectedError), fmt.Sprintf("test %d failed on error returned", i))
	}
}
