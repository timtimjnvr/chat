package crdt

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestGetNodeBytes(t *testing.T) {
	var (
		id, _ = uuid.Parse("4b8e153b-834f-4190-b5d3-aba2f35ead56")

		i = &infos{
			slot:    3,
			Id:      id,
			Name:    "Toto",
			Address: "127.0.0.1",
			Port:    "8080",
		}

		tests = []struct {
			infos         Infos
			expectedBytes []byte
		}{
			{
				infos:         i,
				expectedBytes: []byte{123, 34, 73, 100, 34, 58, 34, 52, 98, 56, 101, 49, 53, 51, 98, 45, 56, 51, 52, 102, 45, 52, 49, 57, 48, 45, 98, 53, 100, 51, 45, 97, 98, 97, 50, 102, 51, 53, 101, 97, 100, 53, 54, 34, 44, 34, 112, 111, 114, 116, 34, 58, 34, 56, 48, 56, 48, 34, 44, 34, 97, 100, 100, 114, 101, 115, 115, 34, 58, 34, 49, 50, 55, 46, 48, 46, 48, 46, 49, 34, 44, 34, 110, 97, 109, 101, 34, 58, 34, 84, 111, 116, 111, 34, 125},
			},
		}
	)

	for i, test := range tests {
		res := test.infos.ToBytes()
		assert.True(t, reflect.DeepEqual(res, test.expectedBytes), fmt.Sprintf("test %d failed on runes returned", i))
	}
}

func TestDecodeNode(t *testing.T) {

	var (
		id, _ = uuid.Parse("4b8e153b-834f-4190-b5d3-aba2f35ead56")

		i = &infos{
			slot:    0,
			Id:      id,
			Name:    "Toto",
			Address: "127.0.0.1",
			Port:    "8080",
		}

		tests = []struct {
			bytes         []byte
			expectedInfos *infos
		}{
			{
				bytes:         []byte{123, 34, 73, 100, 34, 58, 34, 52, 98, 56, 101, 49, 53, 51, 98, 45, 56, 51, 52, 102, 45, 52, 49, 57, 48, 45, 98, 53, 100, 51, 45, 97, 98, 97, 50, 102, 51, 53, 101, 97, 100, 53, 54, 34, 44, 34, 112, 111, 114, 116, 34, 58, 34, 56, 48, 56, 48, 34, 44, 34, 97, 100, 100, 114, 101, 115, 115, 34, 58, 34, 49, 50, 55, 46, 48, 46, 48, 46, 49, 34, 44, 34, 110, 97, 109, 101, 34, 58, 34, 84, 111, 116, 111, 34, 125},
				expectedInfos: i,
			},
		}
	)
	for i, test := range tests {
		res, _ := DecodeInfos(test.bytes)
		assert.True(t, reflect.DeepEqual(res, test.expectedInfos), fmt.Sprintf("test %d failed on struct returned", i))
	}
}
