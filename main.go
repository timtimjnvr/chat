package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		myPortPtr = flag.String("p", "8080", "port number used to accept connections")
		myAddrPtr = flag.String("a", "", "address used to accept connections")
		myNamePtr = flag.String("u", "tim", "nickname used in all chat")

		sigc = make(chan os.Signal, 1)
	)

	flag.Parse()

	signal.Notify(sigc,
		syscall.SIGUSR1, // only used for interruption in testing
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	start(*myAddrPtr, *myPortPtr, *myNamePtr, os.Stdin, sigc)
}
