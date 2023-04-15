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

	assert.True(t, chat.containsMessage(chatMessages[0]))
	assert.True(t, chat.containsMessage(chatMessages[1]))
	assert.True(t, chat.containsMessage(chatMessages[2]))

	var otherMessage = NewMessage("Anonymous", "Hello Guys!")
	assert.False(t, chat.containsMessage(otherMessage))
}

func TestChat_SaveMessage(t *testing.T) {

	chatMessages := make([]*Message, 10)
	chat := NewChat("name")

	// giving message in chronological order and verifying all messages are present
	for i := range chatMessages {
		chatMessages[i] = NewMessage("sender", fmt.Sprintf("%d", i))
		chatMessages[i].Date = fmt.Sprintf("2006-01-02T15:04:0%dZ", i)
		chat.SaveMessage(chatMessages[i])
	}

	for i, m := range chatMessages {
		assert.Equal(t, chat.messages[i].Content, m.Content)
	}

	chat = NewChat("name")
	chatMessages = make([]*Message, 10)

	// giving message in reverse chronological order and verifying all messages are present
	for i := range chatMessages {
		chatMessages[len(chatMessages)-i-1] = NewMessage("sender", fmt.Sprintf("%d", len(chatMessages)-i-1))
		chatMessages[len(chatMessages)-i-1].Date = fmt.Sprintf("2006-01-02T15:04:0%dZ", len(chatMessages)-i-1)
		chat.SaveMessage(chatMessages[len(chatMessages)-i-1])
	}

	for i, m := range chatMessages {
		assert.Equal(t, chat.messages[i].Content, m.Content)
	}

	// inserting message with random dates anbd verifying they are in right order
	chat = NewChat("name")
	for i := 0; i < 10; i++ {
		m := NewMessage("sender", fmt.Sprintf("%d", i))
		m.Date = randomTimestamp().String()
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
