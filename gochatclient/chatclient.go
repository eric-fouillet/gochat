// A chat client library.
// Provides methods to connect to a server and send messages

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/ericfouillet/gochat"
	"github.com/ericfouillet/gochat/gochatutil"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

var host string
var port string

func init() {
	flag.StringVar(&host, "host", "localhost", "The host to connect to")
	flag.StringVar(&port, "port", "8083", "The port to connect to")
}

// ChatClient is a chat client, holding a username, host and port
type ChatClient struct {
	Username   string
	TargetHost string
	TargetPort string
	recv       chan *gochat.ChatMessage
	msgPool    *gochatutil.MsgPool
	connR      *bufio.Reader
	connW      *bufio.Writer
}

// LoginMessage is used for the first login
const LoginMessage string = "GOCHATLOGIN\n"

// Chat client
// Connect to the host and port given in parameters.
// If not provided, the default server URL is localhost:8083.
// After connection, starts reading from stdin some text to send,
// and prints messages received from the server.
func main() {
	flag.Parse()
	fmt.Print("Enter your username: ")
	r := bufio.NewReader(os.Stdin)
	username, err := r.ReadString('\n')
	if err != nil {
		log.Println("Could not read from stdin", err)
		return
	}
	client := &ChatClient{
		Username:   username[:len(username)-1],
		TargetHost: host,
		TargetPort: port,
		recv:       make(chan *gochat.ChatMessage, 10),
		msgPool:    gochatutil.NewPool(),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conn, err := client.connect(ctx)
	if err != nil {
		log.Println("Could not connect to server", err)
	}
	defer conn.Close()

	// Start receiving messages in a separate goroutine
	go client.print(ctx)
	go client.read(ctx, conn)

	// Read from stdin messages to send
	for {
		fmt.Printf("%v> ", client.Username)
		msg, err := r.ReadString('\n')
		if err != nil {
			log.Println("Could not read from stdin", err)
			return
		}
		if msg == "/exit\n" {
			return
		}
		client.send(ctx, msg)
	}
}

// read reads responses from the server
func (cc *ChatClient) read(ctx context.Context, conn net.Conn) {
	for {
		msg, err := cc.doRead(ctx)
		if err != nil {
			if err == io.EOF {
				log.Println("Connection with server was dropped")
				return
			}
			log.Println("Could not read response", err)
		}
		// cc.recv <- msg
		select {
		case cc.recv <- msg:
		case <-ctx.Done():
			return
		}
	}
}

func (cc *ChatClient) print(ctx context.Context) {
	for {
		select {
		case msg := <-cc.recv:
			fmt.Printf("\n%v> %v\n%v> ", msg.GetSender(), msg.GetContent(), cc.Username)
			cc.msgPool.Rel(msg)
		case <-ctx.Done():
			return
		}
	}
}

// Connect to a server running at the given host and port
func (cc *ChatClient) connect(ctx context.Context) (net.Conn, error) {
	addr, err := net.ResolveTCPAddr("tcp", string(cc.TargetHost+":"+cc.TargetPort))
	if err != nil {
		return nil, errors.Wrap(err, "Could not resolve the TCP address")
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, errors.Wrap(err, "Could not establish a connection")
	}
	cc.connR = bufio.NewReader(conn)
	cc.connW = bufio.NewWriter(conn)
	cc.send(ctx, LoginMessage)
	return conn, nil
}

func (cc *ChatClient) send(ctx context.Context, msg string) {
	sendTime := uint64(time.Now().Unix())
	strippedMsg := msg[:len(msg)-1] // remove line return
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
	if _, err := cc.connW.Write(data); err != nil {
		log.Println("Could not write message", err)
		return
	}
	cc.connW.Flush()
}

// Read a message from the server
func (cc *ChatClient) doRead(ctx context.Context) (*gochat.ChatMessage, error) {
	buf := make([]byte, 1024)
	readBytes, err := cc.connR.Read(buf)
	if err != nil {
		return nil, err
	}
	returnMsg := cc.msgPool.Get()
	err = proto.Unmarshal(buf[:readBytes], returnMsg)
	if err != nil {
		return nil, err
	}
	return returnMsg, nil
}
