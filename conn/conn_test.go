package conn

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewConn(t *testing.T) {
	t.Parallel()

	c, _, err := helperGetConnections("12340")
	if err != nil {
		assert.Fail(t, "failed to create a connections")
	}

	_, err = newConn(c)
	assert.NoError(t, err, "newConn return an unexpected error")

	c.Close()
	_, err = newConn(c)
	assert.Error(t, err, "newConn on closed c did not return an error")
}
