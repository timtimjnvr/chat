package storage

import (
	"fmt"
	"github.com/google/uuid"
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
	l.Add(crdt.NewChat("1"))
	l.Add(crdt.NewChat("2"))
	l.Add(crdt.NewChat("3"))

	ass.Equal(3, l.Len(), "failed on Adding elements")
}

func TestList_Contains(t *testing.T) {
	l := NewList()
	id1 := l.Add(crdt.NewChat("1"))
	id2 := l.Add(crdt.NewChat("2"))
	id3 := l.Add(crdt.NewChat("3"))

	errMessage := "failed on finding element"
	assert.True(t, l.Contains(id1), errMessage)
	assert.True(t, l.Contains(id2), errMessage)
	assert.True(t, l.Contains(id3), errMessage)

	id4, _ := uuid.NewUUID()
	assert.False(t, l.Contains(id4), "found non existing element")
}

func TestList_Update(t *testing.T) {
	l := NewList()
	id1 := l.Add(crdt.NewChat("1"))
	l.Update(id1, &crdt.Chat{Id: id1.String(), Name: "3"})
	c, _ := l.GetById(id1)

	assert.Equal(t, &crdt.Chat{Id: id1.String(), Name: "3"}, c, "failed to update chat")
}

func TestList_Delete(t *testing.T) {
	var (
		ass    = assert.New(t)
		l      = NewList()
		first  = l.Add(crdt.NewChat("1"))
		second = l.Add(crdt.NewChat("2"))
		third  = l.Add(crdt.NewChat("3"))
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

	first = l.Add(crdt.NewChat("1"))
	second = l.Add(crdt.NewChat("2"))
	third = l.Add(crdt.NewChat("3"))

	// 1 -> 2 -> 3 becomes 1 -> 3
	l.Delete(second)
	ass.Equal(l.Len(), 2, "failed on Deleting first element")

	// Verify new first = 2
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
}

func TestList_GetById(t *testing.T) {
	var (
		ass       = assert.New(t)
		l         = NewList()
		values    = []int{1, 2, 3, 4}
		valuesIds = make(map[int]uuid.UUID)
	)

	for _, value := range values {
		valuesIds[value] = l.Add(crdt.NewChat(fmt.Sprintf("%d", value)))
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
}
