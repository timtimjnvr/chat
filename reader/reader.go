package reader

import (
	"golang.org/x/sys/unix"
	"log"
	"os"
	"strings"
	"sync"
)

const MaxMessageSize = 1000

func ReadFile(wg *sync.WaitGroup, file *os.File, output chan<- []byte, shutdown chan struct{}) {
	defer func() {
		close(output)
		wg.Done()
		log.Println("[INFO] ReadFile stopped")
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
		log.Println("[INFO] type a command")

		var (
			fdSet  = unix.FdSet{}
			buffer = make([]byte, MaxMessageSize)
			err    error
		)

		fdSet.Clear(int(file.Fd()))
		fdSet.Clear(int(readClose.Fd()))

		fdSet.Set(int(file.Fd()))
		fdSet.Set(int(readClose.Fd()))

		// wait and modifies file descriptors in fdSet with first ready to use file descriptors (ie for us output or readClose)
		_, err = unix.Select(int(readClose.Fd()+1), &fdSet, nil, nil, &unix.Timeval{Sec: 60, Usec: 0})
		if err != nil {
			log.Fatal("[ERROR] ", err)
			return
		}

		// readClose : stop reading output
		if fdSet.IsSet(int(readClose.Fd())) {
			return
		}

		// default : read output
		var n int
		n, err = file.Read(buffer)
		if err != nil {
			return
		}

		if n > 0 {
			if n > 0 {
				splitAndInsertMessages(buffer[:n], output)
			}
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
