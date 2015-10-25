// A chat client library.
// Provides methods to connect to a server and send messages

package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/eric-fouillet/gochat"
	"github.com/eric-fouillet/gochat/gochatutil"
	"github.com/golang/protobuf/proto"
)

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
	host := flag.String("host", "localhost", "The host to connect to")
	port := flag.String("port", "8083", "The port to connect to")
	flag.Parse()
	fmt.Print("Enter your username: ")
	username, errName := bufio.NewReader(os.Stdin).ReadString('\n')
	if gochatutil.CheckError(errName) {
		return
	}
	chatClient := ChatClient{username[:len(username)-1], *host, *port}
	conn, _ := chatClient.Connect(*host, *port)
	defer conn.Close()

	// Start receiving messages in a separate goroutine
	go chatClient.WaitForResponses(conn)

	// Read from stdin messages to send
	for {
		fmt.Printf("%v> ", chatClient.Username)
		msg, err3 := bufio.NewReader(os.Stdin).ReadString('\n')
		if gochatutil.CheckError(err3) {
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
		if !gochatutil.CheckError(err) {
			fmt.Printf("\n%v> %v\n%v> ", msg.GetSender(), msg.GetContent(), cc.Username)
		}
	}
}

// Connect to a server running at the given host and port
func (cc *ChatClient) Connect(host string, port string) (net.Conn, error) {
	addr, err := net.ResolveTCPAddr("tcp", string(host+":"+port))
	if gochatutil.CheckError(err) {
		return nil, err
	}
	conn, err2 := net.DialTCP("tcp", nil, addr)
	if gochatutil.CheckError(err2) {
		return nil, err2
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
	if gochatutil.CheckError(err) {
		return
	}
	_, errWrite := conn.Write(data)
	if gochatutil.CheckError(errWrite) {
		return
	}
}

// Read a message from the server
func (cc *ChatClient) readResponse(conn net.Conn) (*gochat.ChatMessage, error) {
	buf := make([]byte, 1024)
	readBytes, err := bufio.NewReader(conn).Read(buf)
	if gochatutil.CheckError(err) {
		return nil, err
	}
	var returnMsg = new(gochat.ChatMessage)
	err2 := proto.Unmarshal(buf[:readBytes], returnMsg)
	return returnMsg, err2
}
