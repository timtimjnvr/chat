package crdt

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestGetSyncBytes(t *testing.T) {
	id, _ := uuid.NewUUID()
	m := &message{
		id:      id,
		Sender:  "toto",
		Content: "J'aime me battre",
		Date:    time.Now(),
	}

	bytes := m.GetSyncBytes(add)
	log.Println(string(bytes))

	assert.Fail(t, "failed - testing")
}

func TestGetMessageFromBytes(t *testing.T) {
	bytes := []byte{36, 56, 55, 101, 98, 97, 48, 99, 45, 57, 55, 99, 100, 45, 49, 49, 101, 100, 45, 57, 49, 54, 50, 45, 49, 54, 48, 102, 48, 97, 53, 54, 48, 54, 54, 101, 4, 111, 116, 111, 16, 39, 97, 105, 109, 101, 32, 109, 101, 32, 98, 97, 116, 116, 114, 101, 25, 48, 50, 51, 45, 48, 49, 45, 49, 57, 84, 48, 56, 58, 53, 48, 58, 49, 52, 43, 48, 49, 58, 48, 48}
	log.Println(string(bytes))
	m := GetMessageFromBytes(bytes)
	log.Println(m)
	assert.Fail(t, "failed - testing")
}
