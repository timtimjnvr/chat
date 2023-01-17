package crdt

import (
	"chat/node"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"log"
	"reflect"
	"testing"
)

/*
func TestDeserializeOperation(t *testing.T) {
	id, _ := uuid.Parse("3800ad85-7edb-42e4-b399-4a8499bd1768")

	ass := assert.New(t)
	data := make([]message, 1)
	data[0] = message{Sender: "tim", Content: "blabla"}
	var tests = []struct {
		bytesInput        []byte
		expectedOperation Operation
	}{
		{
			bytesInput: []byte("{\"operation\":0,\"data\":[{\"sender\":\"tim\",\"content\":\"blabla\",\"date\":\"0001-01-01T00:00:00Z\"}]}"),
			expectedOperation: messageOperation{
				OperationTypology: add,
				Data: []message{
					{Sender: "tim", Content: "blabla"},
				},
			},
		},
		{
			bytesInput: []byte("{\"operation\":0,\"data\":[{\"id\":\"3800ad85-7edb-42e4-b399-4a8499bd1768\",\"port\":8080,\"address\":\"127.0.0.1\"}]}"),
			expectedOperation: nodeOperation{
				OperationTypology: add,
				Data: []node.Infos{
					{
						Id:      id,
						Port:    8080,
						Address: "127.0.0.1",
					},
				},
			},
		},
	}

	for i, test := range tests {
		op := deserializeOperation(test.bytesInput)
		log.Println("---------")
		log.Println(op)
		log.Println(test.expectedOperation)
		ass.True(reflect.DeepEqual(op, test.expectedOperation), fmt.Sprintf("test %d failed on bytes operation returned", i))
	}
}*/

func TestSerializeOperation(t *testing.T) {
	ass := assert.New(t)
	id, _ := uuid.Parse("3800ad85-7edb-42e4-b399-4a8499bd1768")

	var tests = []struct {
		operation           Operation
		expectedBytesOutput []byte
	}{
		{
			operation: messageOperation{
				OperationTypology: add,
				Data:              []message{{Sender: "tim", Content: "blabla"}},
			},
			expectedBytesOutput: []byte("{\"operation\":0,\"data\":[{\"sender\":\"tim\",\"content\":\"blabla\",\"date\":\"0001-01-01T00:00:00Z\"}]}"),
		},
		{
			operation: nodeOperation{
				OperationTypology: add,
				Data:              []node.Infos{{Id: id, Port: 8080, Address: "127.0.0.1"}},
			},
			expectedBytesOutput: []byte("{\"operation\":0,\"data\":[{\"id\":\"3800ad85-7edb-42e4-b399-4a8499bd1768\",\"port\":8080,\"address\":\"127.0.0.1\"}]}"),
		},
	}

	for i, test := range tests {
		output := serializeOperation(test.operation)
		log.Println(string(output))
		log.Println(string(test.expectedBytesOutput))
		ass.True(reflect.DeepEqual(output, test.expectedBytesOutput), fmt.Sprintf("test %d failed on bytes output returned", i))
	}
}
