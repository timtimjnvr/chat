package conn

import (
	"github.com/pkg/errors"
	"net"
	"os"
)

const defaultRetries = 100

var ErrWriteMaxRetriesReached = errors.New("max retries reached")

type conn struct {
	net.Conn         // used to read, write bytes on the socket
	file     os.File // needed by reader package to get a file descriptor of the socket in unix Select
	retries  int32
}

func newConn(c net.Conn) (*conn, error) {
	file, err := c.(*net.TCPConn).File()
	if err != nil {
		return nil, err
	}

	return &conn{
		Conn:    c,
		file:    *file,
		retries: defaultRetries,
	}, nil
}

func (c *conn) Fd() uintptr {
	return c.file.Fd()
}

func (c *conn) Close() error {
	// close the net.Conn
	c.Conn.Close()
	// close the duplicated file
	return c.file.Close()
}

func (c *conn) Write(message []byte) error {
	var (
		n   int
		err error
	)

	for n != len(message) || c.retries != 0 {
		n, err = c.Conn.Write(message)
		if err == nil {
			return nil
		}
		c.retries--
	}

	if c.retries == 0 {
		return ErrWriteMaxRetriesReached
	}

	return nil
}

func Write(c *conn, message []byte) error {
	var (
		n   int
		err error
	)

	for n != len(message) || c.retries != 0 {
		n, err = c.Conn.Write(message)
		if err == nil {
			return nil
		}
		c.retries--
	}

	if c.retries == 0 {
		return ErrWriteMaxRetriesReached
	}

	return nil
}
