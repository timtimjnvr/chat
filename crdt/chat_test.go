package crdt

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
