package reader

import (
	"bytes"
	"golang.org/x/sys/unix"
	"os"
)

type Reader interface {
	Fd() uintptr
	Read(b []byte) (n int, err error)
	Close() error
}

var Separator = []byte("\n")

const MaxMessageSize = 1000

func Read(reader Reader, output chan<- []byte, separator []byte, shutdown, isDone chan struct{}) {
	done := make(chan struct{})

	defer func() {
		reader.Close()
		close(done)
		close(output)
		close(isDone)
	}()

	// writeClose is closed in order to signal to stop reading output
	var readClose, writeClose, _ = os.Pipe()

	go func(chan struct{}) {
		select {
		case <-shutdown:
			_ = writeClose.Close()
		case <-done:
			return
		}
	}(done)

	for {
		var (
			fdSet  = unix.FdSet{}
			buffer = make([]byte, MaxMessageSize)
			err    error
		)

		// clear fdSet
		fdSet.Zero()
		fdSet.Set(int(reader.Fd()))
		fdSet.Set(int(readClose.Fd()))

		// wait and modifies reader descriptors in fdSet with first ready to use reader descriptors (ie for us reader or readClose)
		someThingToRead, err := unix.Select(int(readClose.Fd()+1), &fdSet, nil, nil, &unix.Timeval{Sec: 5, Usec: 0})
		// nothing to read
		if someThingToRead == 0 {
			continue
		}

		// Interrupted Syscall sometimes
		if err != nil {
			continue
		}

		// readClose : stop reading output
		if fdSet.IsSet(int(readClose.Fd())) {
			return
		}
		// default use reader
		var n int
		n, err = reader.Read(buffer)
		if err != nil {
			return
		}
		if n == 0 {
			return
		}

		// split content into elements and output them
		elements := bytes.Split(buffer[:n], separator)
		for _, element := range elements {
			if len(element) == 0 {
				continue
			}

			output <- element
		}
	}
}
