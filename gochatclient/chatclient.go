package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	args := os.Args[1:]
	addr, err := net.ResolveTCPAddr("tcp", string(args[0]+":"+args[1]))
	if err != nil {
		fmt.Printf("Error while resolving address %#v", err)
		return
	}
	conn, err2 := net.DialTCP("tcp", nil, addr)
	if err2 != nil {
		fmt.Printf("Error while dialing TCP %#v", err2)
		return
	}
	defer conn.Close()
	//readerStd := bufio.NewReader(os.Stdin)
	//readerConn := bufio.NewReader(conn)
	for {
		fmt.Print("Enter text to send: ")
		msg, err3 := bufio.NewReader(os.Stdin).ReadString('\n')
		if err3 != nil {
			fmt.Printf("Error while reading from stdin %v\n", err3)
			continue
		}
		fmt.Fprintf(conn, msg+"\n")
		returnMsg, err4 := bufio.NewReader(conn).ReadString('\n')
		if err4 != nil {
			fmt.Printf("Error while reading from remote ! %v\n", err4)
			continue
		}
		fmt.Printf("Received: %v\n", returnMsg)
	}
}
