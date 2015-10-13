package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/eric-fouillet/gochat"
	"github.com/golang/protobuf/proto"
)

// Chat client
// Connect to the host and port given in parameters
// and start reading from stdin some text to send
func main() {
	host := flag.String("host", "localhost", "The host to connect to")
	port := flag.String("port", "8083", "The port to connect to")
	flag.Parse()
	addr, err := net.ResolveTCPAddr("tcp", string(*host+":"+*port))
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
	for {
		fmt.Print("Enter text to send: ")
		msg, err3 := bufio.NewReader(os.Stdin).ReadString('\n')
		if err3 != nil {
			fmt.Printf("Error while reading from stdin %v\n", err3)
			continue
		}
		if msg == "exit\n" {
			return
		}
		sender := "someone"
		sendTime := uint64(time.Now().Unix())
		strippedMsg := msg[:len(msg)-1]
		protoMsg := &gochat.ChatMessage{
			Sender:   &sender,
			SendTime: &sendTime,
			Content:  &strippedMsg,
		}
		data, err := proto.Marshal(protoMsg)
		if err != nil {
			log.Fatal("marshaling error: ", err)
		}
		_, errWrite := conn.Write(data)
		if errWrite != nil {
			fmt.Printf("Error while writing to remote ! %v\n", errWrite)
			continue
		}
		returnMsg, err4 := bufio.NewReader(conn).ReadString('\n')
		if err4 != nil {
			fmt.Printf("Error while reading from remote ! %v\n", err4)
			continue
		}
		returnStr := returnMsg[:len(returnMsg)-1]
		fmt.Printf("Received: %v\n", returnStr)
	}

}

func checkError(err error, action string) {
	if err != nil {
		log.Fatal("Error %v\n", err)
		switch action {
		case "stop":
			os.Exit(-1)
		case "continue":
			return
		default:
			return
		}
	}
}
