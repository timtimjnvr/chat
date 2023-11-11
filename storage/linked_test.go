package storage

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github/timtimjnvr/chat/crdt"
	"testing"
)

func TestList_Len(t *testing.T) {
	ass := assert.New(t)
	l := NewChatList()
	ass.True(l.Len() == 0, "failed on computing listOld length")
}

func TestList_Add(t *testing.T) {
	ass := assert.New(t)
	l := NewChatList()

	id, err := l.Add(crdt.NewChat("1"))
	assert.Nil(t, err)

	// adding chat with name already in listOld
	existingChat := crdt.NewChat("1")
	_, err = l.Add(existingChat)
	assert.True(t, errors.Is(err, AlreadyInListWithNameErr))

	// adding chat with id already in listOld
	existingChat = crdt.NewChat("toto")
	existingChat.Id = id
	_, err = l.Add(existingChat)
	assert.True(t, errors.Is(err, AlreadyInListWithIDErr))

	// adding more elements
	_, err = l.Add(crdt.NewChat("2"))
	assert.Nil(t, err)
	_, err = l.Add(crdt.NewChat("3"))
	assert.Nil(t, err)

	// checking listOld length
	ass.Equal(3, l.Len(), "failed on Adding elements")
}

func TestList_Contains(t *testing.T) {
	l := NewChatList()
	// test on empty listOld
	inExistingID, _ := uuid.NewUUID()
	contains := l.Contains(inExistingID)
	assert.False(t, contains)

	id1, _ := l.Add(crdt.NewChat("1"))
	id2, _ := l.Add(crdt.NewChat("2"))
	id3, _ := l.Add(crdt.NewChat("3"))

	errMessage := "failed on finding elementOld"
	assert.True(t, l.Contains(id1), errMessage)
	assert.True(t, l.Contains(id2), errMessage)
	assert.True(t, l.Contains(id3), errMessage)

	id4, _ := uuid.NewUUID()
	assert.False(t, l.Contains(id4), "found non existing elementOld")
}

func TestList_Update(t *testing.T) {
	l := NewChatList()
	// Try to update chat in empty listOld
	err := l.Update(crdt.NewChat("non existing"))
	assert.True(t, errors.Is(err, NotFoundErr))

	c := crdt.NewChat("1")
	id1String := c.Id.String()
	l.Add(c)

	// Try to update with invalid chat
	err = l.Update(nil)
	assert.True(t, errors.Is(err, InvalidChatErr))

	err = l.Update(&crdt.Chat{Id: c.Id, Name: "3"})
	// no error
	assert.Nil(t, err)

	id1, err := uuid.Parse(id1String)
	assert.Nil(t, err)
	c, err = l.GetById(id1)
	assert.Nil(t, err)
	assert.Equal(t, &crdt.Chat{Id: id1, Name: "3"}, c, "failed to update chat")
}

func TestList_Delete(t *testing.T) {
	var (
		ass       = assert.New(t)
		l         = NewChatList()
		first, _  = l.Add(crdt.NewChat("1"))
		second, _ = l.Add(crdt.NewChat("2"))
		third, _  = l.Add(crdt.NewChat("3"))
	)
	// 1 -> 2 -> 3 becomes 2 -> 3
	l.Delete(first)
	ass.Equal(l.Len(), 2, "failed on Deleting first elementOld")

	// Verify new first = 2
	c, err := l.GetByIndex(0)
	if err != nil {
		ass.Fail("failed to get elementOld after deleting first")
	}
	ass.Equal(c.Id, second)

	// Verify new second = 3
	c, err = l.GetByIndex(1)
	if err != nil {
		ass.Fail("failed to get elementOld after deleting first")
	}

	ass.Equal(c.Id, third)

	// 2 -> 3 becomes 2
	l.Delete(third)
	ass.Equal(l.Len(), 1, "failed on Deleting third elementOld")

	// Verify new first = 2
	c, err = l.GetByIndex(0)
	if err != nil {
		ass.Fail("failed to get elementOld after deleting first")
	}

	ass.Equal(c.Id, second)

	l.Delete(second)
	ass.Equal(l.Len(), 0, "failed on Deleting remaining elementOld")
	ass.Nil(l.head)

	first, _ = l.Add(crdt.NewChat("1"))
	second, _ = l.Add(crdt.NewChat("2"))
	third, _ = l.Add(crdt.NewChat("3"))

	// 1 -> 2 -> 3 becomes 1 -> 3
	l.Delete(second)
	ass.Equal(l.Len(), 2, "failed on Deleting first elementOld")

	// Verify new first = 1
	c, err = l.GetByIndex(0)
	if err != nil {
		ass.Fail("failed to get elementOld after deleting first")
	}
	ass.Equal(c.Id, first)

	// Verify new second = 3
	c, err = l.GetByIndex(1)
	if err != nil {
		ass.Fail("failed to get elementOld after deleting first")
	}

	ass.Equal(c.Id, third)

	// Try to delete in existing elementOld (2) and validate that nothing has changed
	l.Delete(second)

	// Verify first has not changed = 1
	c, err = l.GetByIndex(0)
	if err != nil {
		ass.Fail("failed to get elementOld after deleting first")
	}
	ass.Equal(c.Id, first)

	// Verify second has not changed = 3
	c, err = l.GetByIndex(1)
	if err != nil {
		ass.Fail("failed to get elementOld after deleting first")
	}

	ass.Equal(c.Id, third)

	// try to delete an elementOld in an empty listOld
	l = NewChatList()
	l.Delete(second)
}

func TestList_GetById(t *testing.T) {
	var (
		ass       = assert.New(t)
		l         = NewChatList()
		values    = []int{1, 2, 3, 4}
		valuesIds = make(map[int]uuid.UUID)
	)

	for _, value := range values {
		valuesIds[value], _ = l.Add(crdt.NewChat(fmt.Sprintf("%d", value)))
	}

	for value, id := range valuesIds {
		res, err := l.GetById(id)
		if err != nil {
			assert.Fail(t, "failed on getting elementOld by id")
			return
		}
		var expectedChat = crdt.NewChat(fmt.Sprintf("%d", value))
		expectedChat.Id = id

		ass.Equal(res, expectedChat, "failed on getting elementOld by id, wrong chat")
	}

	// try to get in existing elementOld in a non empty listOld
	unExistingID, _ := uuid.NewUUID()
	res, err := l.GetById(unExistingID)
	assert.True(t, errors.Is(err, NotFoundErr))
	assert.Nil(t, res)

	// try to get in existing elementOld in an empty listOld
	l = NewChatList()
	res, err = l.GetById(unExistingID)
	assert.True(t, errors.Is(err, NotFoundErr))
	assert.Nil(t, res)
}

func TestList_GetByIndex(t *testing.T) {
	l := NewChatList()

	// Try to get more than length
	e, err := l.GetByIndex(2)
	assert.True(t, errors.Is(err, NotFoundErr))
	assert.Nil(t, e)

	var (
		values = []int{1, 2, 3, 4}
	)

	for _, value := range values {
		l.Add(crdt.NewChat(fmt.Sprintf("%d", value)))
	}

	for index, value := range values {
		var c *crdt.Chat
		c, err = l.GetByIndex(index)
		assert.NoError(t, err)
		assert.Equal(t, c.Name, fmt.Sprintf("%d", value))
	}
}
