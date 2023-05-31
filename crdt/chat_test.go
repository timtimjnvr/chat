package crdt

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestContainsMessage(t *testing.T) {
	var (
		chatMessages = []*Message{
			NewMessage("James", "Hello Bob!"),
			NewMessage("Bob", "Hello James!"),
			NewMessage("Peter", "Hello James!"),
		}

		chat = Chat{
			messages: chatMessages,
		}
	)

	assert.True(t, chat.ContainsMessage(chatMessages[0]))
	assert.True(t, chat.ContainsMessage(chatMessages[1]))
	assert.True(t, chat.ContainsMessage(chatMessages[2]))

	var otherMessage = NewMessage("Anonymous", "Hello Guys!")
	assert.False(t, chat.ContainsMessage(otherMessage))
}

func TestChat_SaveMessage(t *testing.T) {
	// inserting message with random dates and verifying they are in right order
	chat := NewChat("name")
	for i := 0; i < 10; i++ {
		m := NewMessage("sender", fmt.Sprintf("%d", i))
		m.Date = randomTimestamp().Format(time.RFC3339)
		chat.SaveMessage(m)
	}

	var (
		currentDate     time.Time
		previousDate, _ = time.Parse(time.RFC3339, chat.messages[0].Date)
	)
	for i := 1; i < len(chat.messages); i++ {
		currentDate, _ = time.Parse(time.RFC3339, chat.messages[i].Date)
		assert.True(t, currentDate.After(previousDate))
	}
}

func randomTimestamp() time.Time {
	randomTime := rand.Int63n(time.Now().Unix()-94608000) + 94608000

	randomNow := time.Unix(randomTime, 0)

	return randomNow
}

func TestEncodeChatFields(t *testing.T) {
	var (
		uuidString = "4b8e153b-834f-4190-b5d3-aba2f35ead56"
	)

	var (
		tests = []struct {
			chatToEncode  *Chat
			chatExpected  *Chat
			expectedError error
		}{
			{
				chatToEncode: &Chat{
					Id:   uuidString,
					Name: "James",
					nodesInfos: []*NodeInfos{
						{
							Name: "test",
						},
					},
					messages: []*Message{
						{
							Sender: "test",
						},
					},
				},
				chatExpected: &Chat{
					Id:   uuidString,
					Name: "James",
				},
				expectedError: nil,
			},
		}
	)

	for i, test := range tests {
		bytes := test.chatToEncode.ToBytes()
		var res = &Chat{}
		err := decodeData(bytes, res)
		assert.Equal(t, test.chatExpected, res)
		assert.True(t, reflect.DeepEqual(err, test.expectedError), fmt.Sprintf("test %d failed on error returned", i))
	}
}
