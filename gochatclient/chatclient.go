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

// Chat client
// Connect to the host and port given in parameters
// and start reading from stdin some text to send
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
	for {
		fmt.Print("Enter text to send: ")
		msg, err3 := bufio.NewReader(os.Stdin).ReadString('\n')
		if gochatutil.CheckError(err3) {
			return
		}
		if msg == "exit\n" {
			return
		}
		chatClient.SendMessage(msg, conn)
		chatClient.ReadResponse(conn)
	}
}

func (cc *ChatClient) Connect(host string, port string) (net.Conn, error) {
	addr, err := net.ResolveTCPAddr("tcp", string(host+":"+port))
	if gochatutil.CheckError(err) {
		return nil, err
	}
	conn, err2 := net.DialTCP("tcp", nil, addr)
	if gochatutil.CheckError(err2) {
		return nil, err2
	}
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

func (cc *ChatClient) ReadResponse(conn net.Conn) {
	buf := make([]byte, 1024)
	readBytes, err := bufio.NewReader(conn).Read(buf)
	if gochatutil.CheckError(err) {
		return
	}
	var returnMsg = new(gochat.ChatMessage)
	err2 := proto.Unmarshal(buf[:readBytes], returnMsg)
	if gochatutil.CheckError(err2) {
		return
	}
	fmt.Printf("Received: %v\n", returnMsg.GetContent())
}
