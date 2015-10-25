package main

import (
	"bufio"
	"bytes"
	"container/list"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/eric-fouillet/gochat"
	"github.com/eric-fouillet/gochat/gochatutil"
	"github.com/golang/protobuf/proto"
)

// ChatServer represents a chat server
type ChatServer struct {
	Address  *net.TCPAddr
	Name     string
	BindAddr net.TCPAddr
	Clients  *list.List // List of ServerClient
	Messages chan gochat.ChatMessage
}

// A client connected to the server
type ServerClient struct {
	Conn     net.Conn
	Username string
	Messages chan gochat.ChatMessage
}

// Checks whether 2 Clients are equal
func (c *ServerClient) Equal(other *ServerClient) bool {
	return bytes.Equal([]byte(c.Username), []byte(other.Username)) && c.Conn == other.Conn
}

// Message used for the first login
const LOGIN_MESSAGE string = "GOCHATLOGIN"

// Transform a host and port into a net.TCPAddr
func (cs *ChatServer) parseAddress(addr string, port string) (*net.TCPAddr, error) {
	var addrBuf bytes.Buffer
	if addr != "" {
		addrBuf.WriteString(addr)
	}
	addrBuf.WriteString(":")
	addrBuf.WriteString(port)
	return net.ResolveTCPAddr("tcp", addrBuf.String())
}

// Starts listening on the address and port given in parameters
func (cs *ChatServer) Start(addr string, port string) (err error) {
	cs.Address, err = cs.parseAddress(addr, port)
	gochatutil.CheckError(err)
	cs.Clients = list.New()
	cs.Messages = make(chan gochat.ChatMessage)
	listener, err2 := net.ListenTCP("tcp", cs.Address)
	gochatutil.CheckError(err2)
	log.Printf("Listening on address %v\n", cs.Address.String())
	go cs.DispatchMessages()
	for {
		log.Println("Waiting for connections ...")
		conn, _ := listener.Accept()
		log.Printf("Accepted connection from %v !\n", conn.RemoteAddr().String())
		go cs.handleConnection(conn)
	}
}

// handleConnection: reads requests from clients and sends responses
func (cs *ChatServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		receivedMsg, err := cs.ReadMessage(conn)
		if gochatutil.CheckError(err) {
			break
		}
		log.Println("Starting goroutines")
		go cs.ReceiveMessages(conn)
		client := cs.newClient(receivedMsg, conn)
		go client.SendMessages()
	}
	log.Printf("Closing connection from %v\n", conn.RemoteAddr())
}

// Receive a message on a given connection, and enqueue it
func (cs *ChatServer) ReceiveMessages(conn net.Conn) {
	log.Println("Waiting from messages to receive")
	for {
		receivedMsg, err := cs.ReadMessage(conn)
		if gochatutil.CheckError(err) {
			break
		}
		log.Println("Received a message: add it to the server channel")
		cs.Messages <- *receivedMsg
	}
}

// Dispatch received messages to the appropriate clients
func (cs *ChatServer) DispatchMessages() {
	for {
		log.Println("Waiting for messages to dispatch")
		receiveMsg := <-cs.Messages
		log.Println("Took a message from the server channel, and send it to other clients")
		for current := cs.Clients.Front(); current != nil; current = current.Next() {
			currentClient := current.Value.(ServerClient)
			if currentClient.Username != receiveMsg.GetSender() {
				currentClient.Messages <- receiveMsg
			}
		}
	}
}

// Send a message on a client connection
func (sc *ServerClient) SendMessages() {
	log.Println("Waiting for messages to send")
	msg := <-sc.Messages
	log.Println("Take a message from the client channel, and send it")
	sc.SendMessage(msg.GetSender(), msg.GetContent())
}

// setClient: Adds a client to a chat server
// Also notifies other users that this user has connected
func (cs *ChatServer) newClient(loginMsg *gochat.ChatMessage, conn net.Conn) *ServerClient {
	from := loginMsg.GetSender()
	log.Printf("Adding new client %v to client list\n", from)
	newClient := &ServerClient{conn, from, make(chan gochat.ChatMessage)}
	cs.Clients.PushBack(newClient)
	msg := fmt.Sprintf("User %v has connected", from)
	sendTime := uint64(time.Now().Unix())
	var response = gochat.ChatMessage{
		Sender:   &from,
		SendTime: &sendTime,
		Content:  &msg,
	}
	log.Printf("Adding message %v to the server queue", msg)
	cs.Messages <- response
	log.Println("AFTER")
	return newClient
}

func (cs *ChatServer) ReadMessage(conn net.Conn) (*gochat.ChatMessage, error) {
	buf := make([]byte, 1024)
	readBytes, err := bufio.NewReader(conn).Read(buf)
	if gochatutil.CheckError(err) {
		return nil, err
	}
	var receivedMsg = new(gochat.ChatMessage)
	err2 := proto.Unmarshal(buf[:readBytes], receivedMsg)
	if gochatutil.CheckError(err2) {
		return nil, err
	}
	log.Printf("Read: %v from %v on %v\n", receivedMsg.GetContent(), receivedMsg.GetSender(), conn.RemoteAddr())
	return receivedMsg, nil
}

func (sc *ServerClient) SendMessage(from string, msg string) {
	sendTime := uint64(time.Now().Unix())
	var response = gochat.ChatMessage{
		Sender:   &from,
		SendTime: &sendTime,
		Content:  &msg,
	}
	msgBytes, err3 := proto.Marshal(&response)
	if gochatutil.CheckError(err3) {
		return
	}
	_, err4 := sc.Conn.Write(msgBytes)
	if gochatutil.CheckError(err4) {
		return
	}
}

func main() {
	host := flag.String("host", "localhost", "The host to listen on")
	port := flag.String("port", "8083", "The port to listen on")
	flag.Parse()
	cs := ChatServer{}
	cs.Start(*host, *port)
}
