package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	protocol    = "tcp"
	ip          = ""
	port        = 8080
	addressFmrt = "%s:%d"
)

func main() {
	err := connect(protocol, ip, port)
	if err != nil {
		log.Fatal(err)
	}
}

func connect(protocol, ip string, port int) error {
	conn, err := net.Dial(protocol, fmt.Sprintf(addressFmrt, ip, port))
	if err != nil {
		return err
	}
	r := bufio.NewReader(os.Stdin)
	log.Print("Enter text: ")
	text, _ := r.ReadString('\n')
	buffer := []byte(text)
	_, err = conn.Write(buffer)
	if err != nil {
		return err
	}
	return nil
}
