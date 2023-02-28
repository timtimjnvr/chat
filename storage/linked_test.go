package storage

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLen(t *testing.T) {
	ass := assert.New(t)
	l := NewList()
	ass.True(l.Len() == 0, "failed on computing list length")
}

func TestAdd(t *testing.T) {
	ass := assert.New(t)
	l := NewList()
	l.Add(1)
	l.Add(2)
	l.Add(3)

	ass.Equal(3, l.Len(), "failed on Adding elements")
}

func TestDelete(t *testing.T) {
	var (
		ass    = assert.New(t)
		l      = NewList()
		first  = l.Add(1)
		second = l.Add(2)
		third  = l.Add(3)
	)

	l.Delete(first)
	ass.True(l.Len() == 2, "failed on Deleting first element")

	l.Delete(third)
	ass.True(l.Len() == 1, "failed on Deleting third element")

	l.Delete(second)
	ass.True(l.Len() == 0, "failed on Deleting remaining element")
}

func TestList_GetById(t *testing.T) {
	var (
		ass       = assert.New(t)
		l         = NewList()
		values    = []int{1, 2, 3, 4}
		valuesIds = make(map[int]uuid.UUID)
	)

	for _, value := range values {
		valuesIds[value] = l.Add(value)
	}

	for value, id := range valuesIds {
		res, err := l.GetById(id)
		if err != nil {
			assert.Fail(t, "failed on getting element by id")
			return
		}

		ass.Equal(res.(int), value, "failed on getting element by id, wrong value")
	}
}
