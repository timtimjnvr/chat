package main

import (
	"flag"
	"os"
)

func main() {
	var (
		myPortPtr = flag.String("p", "8080", "port number used to accept connections")
		myAddrPtr = flag.String("a", "", "address used to accept connections")
		myNamePtr = flag.String("u", "tim", "nickname used in all chat")
	)

	flag.Parse()

	start(*myAddrPtr, *myPortPtr, *myNamePtr, os.Stdin)
}
