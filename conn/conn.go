package conn

import (
	"fmt"
	"net"
	"os"
)

type conn struct {
	net.Conn         // used to read, write bytes on the socket
	file     os.File // needed by reader package to get a file descriptor of the socket in unix Select
}

func newConn(c net.Conn) (*conn, error) {
	f, ok := c.(*net.TCPConn)
	if !ok {
		return nil, fmt.Errorf("conn can't be converted to *net.TCPConn")
	}

	file, err := f.File()
	if err != nil {
		return nil, err
	}
	return &conn{
		Conn: c,
		file: *file,
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
