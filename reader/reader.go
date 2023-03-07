package reader

import (
	"golang.org/x/sys/unix"
	"log"
	"os"
	"strings"
	"sync"
)

type Reader interface {
	Fd() uintptr
	Read(b []byte) (n int, err error)
	Close() error
}

const MaxMessageSize = 1000

func Read(wg *sync.WaitGroup, reader Reader, output chan<- []byte, shutdown chan struct{}) {
	defer func() {
		reader.Close()
		close(output)
		wg.Done()
	}()

	// writeClose is closed in order to signal to stop reading output
	var readClose, writeClose, _ = os.Pipe()

	go func() {
		select {
		case <-shutdown:
			_ = writeClose.Close()
		}
	}()

	for {
		var (
			fdSet  = unix.FdSet{}
			buffer = make([]byte, MaxMessageSize)
			err    error
		)

		fdSet.Clear(int(reader.Fd()))
		fdSet.Clear(int(readClose.Fd()))

		fdSet.Set(int(reader.Fd()))
		fdSet.Set(int(readClose.Fd()))

		// wait and modifies reader descriptors in fdSet with first ready to use reader descriptors (ie for us reader or readClose)
		_, err = unix.Select(int(readClose.Fd()+1), &fdSet, nil, nil, &unix.Timeval{Sec: 3600, Usec: 0})
		if err != nil {
			log.Fatal("[ERROR] ", err)
			return
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

		if n > 0 {
			splitAndInsertMessages(buffer[:n], output)
		}
	}
}

func splitAndInsertMessages(buffer []byte, messages chan<- []byte) {
	splitMessages := strings.Split(string(buffer), "\n")
	for _, m := range splitMessages {
		if m != "" {
			messages <- []byte(m)
		}
	}
}
