package crdt

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
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
