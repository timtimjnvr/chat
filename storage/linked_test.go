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
	l := NewList()
	ass.True(l.Len() == 0, "failed on computing list length")
}

func TestList_Add(t *testing.T) {
	ass := assert.New(t)
	l := NewList()

	id, err := l.Add(crdt.NewChat("1"))
	assert.Nil(t, err)

	// adding chat with name already in list
	existingChat := crdt.NewChat("1")
	_, err = l.Add(existingChat)
	assert.True(t, errors.Is(err, AlreadyInListWithNameErr))

	// adding chat with id already in list
	existingChat = crdt.NewChat("toto")
	existingChat.Id = id.String()
	_, err = l.Add(existingChat)
	assert.True(t, errors.Is(err, AlreadyInListWithIDErr))

	// adding more elements
	_, err = l.Add(crdt.NewChat("2"))
	assert.Nil(t, err)
	_, err = l.Add(crdt.NewChat("3"))
	assert.Nil(t, err)

	// checking list length
	ass.Equal(3, l.Len(), "failed on Adding elements")
}

func TestList_Contains(t *testing.T) {
	l := NewList()
	// test on empty list
	inExistingID, _ := uuid.NewUUID()
	contains := l.Contains(inExistingID)
	assert.False(t, contains)

	id1, _ := l.Add(crdt.NewChat("1"))
	id2, _ := l.Add(crdt.NewChat("2"))
	id3, _ := l.Add(crdt.NewChat("3"))

	errMessage := "failed on finding element"
	assert.True(t, l.Contains(id1), errMessage)
	assert.True(t, l.Contains(id2), errMessage)
	assert.True(t, l.Contains(id3), errMessage)

	id4, _ := uuid.NewUUID()
	assert.False(t, l.Contains(id4), "found non existing element")
}

func TestList_Update(t *testing.T) {
	l := NewList()
	// Try to update chat in empty list
	err := l.Update(crdt.NewChat("non existing"))
	assert.True(t, errors.Is(err, NotFoundErr))

	c := crdt.NewChat("1")
	id1String := c.Id
	l.Add(c)

	// Try to update with invalid chat
	err = l.Update(nil)
	assert.True(t, errors.Is(err, InvalidChatErr))

	// Try to update with chat with invalid identifier
	c2 := crdt.NewChat("invalid")
	c2.Id = "toto"
	err = l.Update(c2)
	assert.True(t, errors.Is(err, InvalidIdentifierErr))

	err = l.Update(&crdt.Chat{Id: id1String, Name: "3"})
	// no error
	assert.Nil(t, err)

	id1, err := uuid.Parse(id1String)
	assert.Nil(t, err)
	c, err = l.GetById(id1)
	assert.Nil(t, err)
	assert.Equal(t, &crdt.Chat{Id: id1String, Name: "3"}, c, "failed to update chat")
}

func TestList_Delete(t *testing.T) {
	var (
		ass       = assert.New(t)
		l         = NewList()
		first, _  = l.Add(crdt.NewChat("1"))
		second, _ = l.Add(crdt.NewChat("2"))
		third, _  = l.Add(crdt.NewChat("3"))
	)
	// 1 -> 2 -> 3 becomes 2 -> 3
	l.Delete(first)
	ass.Equal(l.Len(), 2, "failed on Deleting first element")

	// Verify new first = 2
	c, err := l.GetByIndex(0)
	if err != nil {
		ass.Fail("failed to get element after deleting first")
	}
	ass.Equal(c.Id, second.String())

	// Verify new second = 3
	c, err = l.GetByIndex(1)
	if err != nil {
		ass.Fail("failed to get element after deleting first")
	}

	ass.Equal(c.Id, third.String())

	// 2 -> 3 becomes 2
	l.Delete(third)
	ass.Equal(l.Len(), 1, "failed on Deleting third element")

	// Verify new first = 2
	c, err = l.GetByIndex(0)
	if err != nil {
		ass.Fail("failed to get element after deleting first")
	}

	ass.Equal(c.Id, second.String())

	l.Delete(second)
	ass.Equal(l.Len(), 0, "failed on Deleting remaining element")

	first, _ = l.Add(crdt.NewChat("1"))
	second, _ = l.Add(crdt.NewChat("2"))
	third, _ = l.Add(crdt.NewChat("3"))

	// 1 -> 2 -> 3 becomes 1 -> 3
	l.Delete(second)
	ass.Equal(l.Len(), 2, "failed on Deleting first element")

	// Verify new first = 1
	c, err = l.GetByIndex(0)
	if err != nil {
		ass.Fail("failed to get element after deleting first")
	}
	ass.Equal(c.Id, first.String())

	// Verify new second = 3
	c, err = l.GetByIndex(1)
	if err != nil {
		ass.Fail("failed to get element after deleting first")
	}

	ass.Equal(c.Id, third.String())

	// Try to delete in existing element (2) and validate that nothing has changed
	l.Delete(second)

	// Verify first has not changed = 1
	c, err = l.GetByIndex(0)
	if err != nil {
		ass.Fail("failed to get element after deleting first")
	}
	ass.Equal(c.Id, first.String())

	// Verify second has not changed = 3
	c, err = l.GetByIndex(1)
	if err != nil {
		ass.Fail("failed to get element after deleting first")
	}

	ass.Equal(c.Id, third.String())

	// try to delete an element in an empty list
	l = NewList()
	l.Delete(second)
}

func TestList_GetById(t *testing.T) {
	var (
		ass       = assert.New(t)
		l         = NewList()
		values    = []int{1, 2, 3, 4}
		valuesIds = make(map[int]uuid.UUID)
	)

	for _, value := range values {
		valuesIds[value], _ = l.Add(crdt.NewChat(fmt.Sprintf("%d", value)))
	}

	for value, id := range valuesIds {
		res, err := l.GetById(id)
		if err != nil {
			assert.Fail(t, "failed on getting element by id")
			return
		}
		var expectedChat = crdt.NewChat(fmt.Sprintf("%d", value))
		expectedChat.Id = id.String()

		ass.Equal(res, expectedChat, "failed on getting element by id, wrong chat")
	}

	// try to get in existing element in a non empty list
	unExistingID, _ := uuid.NewUUID()
	res, err := l.GetById(unExistingID)
	assert.True(t, errors.Is(err, NotFoundErr))
	assert.Nil(t, res)

	// try to get in existing element in an empty list
	l = NewList()
	res, err = l.GetById(unExistingID)
	assert.True(t, errors.Is(err, NotFoundErr))
	assert.Nil(t, res)
}

func TestList_GetByIndex(t *testing.T) {
	l := NewList()

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
