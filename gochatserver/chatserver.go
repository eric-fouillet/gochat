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
		defer conn.Close()
		log.Printf("Accepted connection from %v !\n", conn.RemoteAddr().String())
		go cs.handleConnection(conn)
	}
}

// handleConnection: reads requests from clients and sends responses
func (cs *ChatServer) handleConnection(conn net.Conn) {
	// Receive the first login message
	receivedMsg, err := cs.ReadMessage(conn)
	if gochatutil.CheckError(err) {
		return
	}
	// Add the client to the list of clients
	client := cs.newClient(receivedMsg, conn)
	cs.printClients()
	go cs.ReceiveMessages(conn)
	go client.SendMessages()
}

func (cs *ChatServer) printClients() {
	i := 0
	log.Println("List of connected clients")
	log.Println("======================================")
	for curr := cs.Clients.Front(); curr != nil; curr = curr.Next() {
		client := curr.Value.(*ServerClient)
		log.Printf("Client %v: %v (%v)\n", i, client.Username, client.Conn.RemoteAddr())
		i++
	}
	log.Println("======================================")
}

// Receive a message on a given connection, and enqueue it
func (cs *ChatServer) ReceiveMessages(conn net.Conn) {
	for {
		receivedMsg, err := cs.ReadMessage(conn)
		if gochatutil.CheckError(err) {
			break
		}
		log.Println("ReceiveMessages: Received a message: add it to the server channel")
		cs.Messages <- *receivedMsg
	}
}

// Dispatch received messages to the appropriate clients
func (cs *ChatServer) DispatchMessages() {
	for {
		receiveMsg := <-cs.Messages
		for current := cs.Clients.Front(); current != nil; current = current.Next() {
			currentClient := current.Value.(*ServerClient)
			if currentClient.Username != receiveMsg.GetSender() {
				log.Printf("Dispatch: sending to client : %v(%v)\n", currentClient.Username, currentClient.Conn.RemoteAddr())
				currentClient.Messages <- receiveMsg
			}
		}
	}
}

// Send a message on a client connection
func (sc *ServerClient) SendMessages() {
	for {
		msg := <-sc.Messages
		log.Printf("(%v) SendMessages: Took a message from the client channel, and will send it\n", sc.Username)
		sc.SendMessage(msg.GetSender(), msg.GetContent())
	}
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
	cs.Messages <- response
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
	log.Printf("(%v) Sending message from %v: [%v]", sc.Username, from, msg)
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
