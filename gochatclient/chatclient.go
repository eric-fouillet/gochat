// A chat client library.
// Provides methods to connect to a server and send messages

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/ericfouillet/gochat"
	"github.com/golang/protobuf/proto"
)

var host string
var port string

func init() {
	flag.StringVar(&host, "host", "localhost", "The host to connect to")
	flag.StringVar(&port, "port", "8083", "The port to connect to")
}

// A chat client, holding a username, host and port
type ChatClient struct {
	Username   string
	TargetHost string
	TargetPort string
}

// Message used for the first login
const LOGIN_MESSAGE string = "GOCHATLOGIN\n"

// Chat client
// Connect to the host and port given in parameters.
// If not provided, the default server URL is localhost:8083.
// After connection, starts reading from stdin some text to send,
// and prints messages received from the server.
func main() {
	flag.Parse()
	fmt.Print("Enter your username: ")
	username, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		log.Println("Could not read from stdin", err)
		return
	}
	chatClient := ChatClient{username[:len(username)-1], host, port}
	conn, _ := chatClient.Connect(host, port)
	defer conn.Close()

	// Start receiving messages in a separate goroutine
	go chatClient.WaitForResponses(conn)

	// Read from stdin messages to send
	for {
		fmt.Printf("%v> ", chatClient.Username)
		msg, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Println("Could not read from stdin", err)
			return
		}
		if msg == "/exit\n" {
			return
		}
		chatClient.SendMessage(msg, conn)
	}
}

// Read responses from the server
func (cc *ChatClient) WaitForResponses(conn net.Conn) {
	for {
		msg, err := cc.readResponse(conn)
		if err != nil {
			if err == io.EOF {
				log.Println("Connection with server was dropped")
				//TODO initialize a reconnection process
				return
			} else {
				log.Println("Could not read response", err)
			}
		} else {
			fmt.Printf("\n%v> %v\n%v> ", msg.GetSender(), msg.GetContent(), cc.Username)
		}
	}
}

// Connect to a server running at the given host and port
func (cc *ChatClient) Connect(host string, port string) (net.Conn, error) {
	addr, err := net.ResolveTCPAddr("tcp", string(host+":"+port))
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}
	cc.SendMessage(LOGIN_MESSAGE, conn)
	return conn, nil
}

func (cc *ChatClient) SendMessage(msg string, conn net.Conn) {
	sendTime := uint64(time.Now().Unix())
	strippedMsg := msg[:len(msg)-1]
	protoMsg := &gochat.ChatMessage{
		Sender:   &cc.Username,
		SendTime: &sendTime,
		Content:  &strippedMsg,
	}
	data, err := proto.Marshal(protoMsg)
	if err != nil {
		log.Println("Could not marshal message", err)
		return
	}
	if _, err := conn.Write(data); err != nil {
		log.Println("Could not write message", err)
		return
	}
}

// Read a message from the server
func (cc *ChatClient) readResponse(conn net.Conn) (*gochat.ChatMessage, error) {
	buf := make([]byte, 1024)
	readBytes, err := bufio.NewReader(conn).Read(buf)
	if err != nil {
		return nil, err
	}
	var returnMsg = new(gochat.ChatMessage)
	err2 := proto.Unmarshal(buf[:readBytes], returnMsg)
	return returnMsg, err2
}
