package main

import (
	"bufio"
	"container/list"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/ericfouillet/gochat"
	"github.com/ericfouillet/gochat/gochatutil"
	"github.com/golang/protobuf/proto"
)

var host string
var port string

func init() {
	flag.StringVar(&host, "host", "localhost", "The host to listen on")
	flag.StringVar(&port, "port", "8083", "The port to listen on")
}

// ChatServer represents a chat server
type ChatServer struct {
	Address  string
	Name     string
	BindAddr net.TCPAddr
	Clients  *list.List // List of ServerClient
	Messages chan *gochat.ChatMessage
	msgPool  *gochatutil.MsgPool
}

// A client connected to the server
type serverClient struct {
	writer     *bufio.Writer
	username   string
	msgs       chan gochat.ChatMessage
	remoteAddr string
}

// LoginMessage is a message used for the first login.
const LoginMessage string = "GOCHATLOGIN"

// start starts listening on the address and port given in parameters
func (cs *ChatServer) start(addr string, port string) (err error) {
	cs.Address = addr + ":" + port
	tcpAddr, err := net.ResolveTCPAddr("tcp", cs.Address)
	if err != nil {
		log.Fatal("Could not resolve address", cs.Address)
	}
	cs.Clients = list.New()
	cs.Messages = make(chan *gochat.ChatMessage)
	cs.msgPool = gochatutil.NewPool()
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatal("Could not listen on address ", cs.Address)
	}
	log.Printf("Listening on address %v\n", cs.Address)
	go cs.dispatchMessages()
	for {
		log.Println("Waiting for connections ...")
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Unable to accept connections on address", cs.Address)
		}
		defer conn.Close()
		log.Printf("Accepted connection from %v !\n", conn.RemoteAddr().String())
		go cs.handleConnection(conn)
	}
}

// handleConnection: reads requests from clients and sends responses
func (cs *ChatServer) handleConnection(conn net.Conn) {
	// Receive the first login message
	log.Println("Waiting for login message")
	receivedMsg, err := cs.read(conn)
	if err != nil {
		log.Println("Error while reading message", err)
		return
	}
	// Add the client to the list of clients
	client := cs.newClient(receivedMsg, conn)
	cs.printClients()
	go cs.receive(conn)
	go client.send()
}

func (cs *ChatServer) printClients() {
	i := 0
	log.Println("List of connected clients")
	log.Println("======================================")
	for curr := cs.Clients.Front(); curr != nil; curr = curr.Next() {
		client := curr.Value.(*serverClient)
		log.Printf("Client %v: %v (%v)\n", i, client.username, client.remoteAddr)
		i++
	}
	log.Println("======================================")
}

// Receive a message on a given connection, and enqueue it
func (cs *ChatServer) receive(conn net.Conn) {
	for {
		receivedMsg, err := cs.read(conn)
		if err != nil {
			log.Println("Could not process message on connection", conn)
			return
		}
		log.Println("ReceiveMessages: Received a message: add it to the server channel")
		cs.Messages <- receivedMsg
	}
}

// Dispatch received messages to the appropriate clients
func (cs *ChatServer) dispatchMessages() {
	for {
		receiveMsg := <-cs.Messages
		for current := cs.Clients.Front(); current != nil; current = current.Next() {
			currentClient := current.Value.(*serverClient)
			if currentClient.username != receiveMsg.GetSender() {
				log.Printf("Dispatch: sending to client : %v(%v)\n", currentClient.username, currentClient.remoteAddr)
				currentClient.msgs <- *receiveMsg
			}
		}
		cs.msgPool.Rel(receiveMsg)
	}
}

// Send a message on a client connection
func (sc *serverClient) send() {
	for {
		msg := <-sc.msgs
		log.Printf("(%v) SendMessages: Took a message from the client channel, and will send it\n", sc.username)
		sc.doSend(msg.GetSender(), msg.GetContent())
	}
}

// setClient: Adds a client to a chat server
// Also notifies other users that this user has connected
func (cs *ChatServer) newClient(loginMsg *gochat.ChatMessage, conn net.Conn) *serverClient {
	from := loginMsg.GetSender()
	log.Printf("Adding new client %v to client list\n", from)
	newClient := &serverClient{
		writer:     bufio.NewWriter(conn),
		username:   from,
		msgs:       make(chan gochat.ChatMessage),
		remoteAddr: conn.RemoteAddr().String(),
	}
	cs.Clients.PushBack(newClient)
	msg := fmt.Sprintf("User %v has connected", from)
	sendTime := uint64(time.Now().Unix())
	var response = &gochat.ChatMessage{
		Sender:   &from,
		SendTime: &sendTime,
		Content:  &msg,
	}
	cs.Messages <- response
	return newClient
}

func (cs *ChatServer) read(conn net.Conn) (*gochat.ChatMessage, error) {
	buf := make([]byte, 1024)
	readBytes, err := bufio.NewReader(conn).Read(buf)
	if err != nil {
		return nil, err
	}
	var receivedMsg = cs.msgPool.Get()
	if err := proto.Unmarshal(buf[:readBytes], receivedMsg); err != nil {
		return nil, err
	}
	log.Printf("Read: %v from %v on %v\n", receivedMsg.GetContent(), receivedMsg.GetSender(), conn.RemoteAddr())
	return receivedMsg, nil
}

func (sc *serverClient) doSend(from string, msg string) {
	log.Printf("(%v) Sending message from %v: [%v]", sc.username, from, msg)
	sendTime := uint64(time.Now().Unix())
	var response = gochat.ChatMessage{
		Sender:   &from,
		SendTime: &sendTime,
		Content:  &msg,
	}
	msgBytes, err := proto.Marshal(&response)
	if err != nil {
		log.Println("Could not marshal response", err)
		return
	}
	if _, err := sc.writer.Write(msgBytes); err != nil {
		log.Printf("Could not write message %v from %v", msg, from)
		return
	}
	sc.writer.Flush()
}

func main() {
	flag.Parse()
	cs := ChatServer{}
	cs.start(host, port)
}
