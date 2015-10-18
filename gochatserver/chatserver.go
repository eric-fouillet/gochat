package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net"
	"strings"
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
	Clients  []net.Conn
}

// Starts listening on the address and port given in parameters
func (cs *ChatServer) start(addr string, port string) (err error) {
	cs.Address, err = parseAddress(addr, port)
	gochatutil.CheckError(err)
	listener, err2 := net.ListenTCP("tcp", cs.Address)
	gochatutil.CheckError(err2)
	fmt.Printf("Listening on address %v\n", cs.Address.String())
	for {
		conn, _ := listener.Accept()
		cs.addClient(conn)
		fmt.Printf("Accepted connection from %v !\n", conn.RemoteAddr().String())
		go cs.handleConnection(conn)
	}
}

// addClient: Adds a client to a chat server
func (cs *ChatServer) addClient(conn net.Conn) {
	cs.Clients = append(cs.Clients, conn)
}

func (cs *ChatServer) ReadMessage(conn net.Conn) (gochat.ChatMessage, error) {
	buf := make([]byte, 1024)
	readBytes, err := bufio.NewReader(conn).Read(buf)
	if gochatutil.CheckError(err) {
		return "", err
	}
	var receivedMsg = new(gochat.ChatMessage)
	err2 := proto.Unmarshal(buf[:readBytes], receivedMsg)
	if gochatutil.CheckError(err2) {
		return "", err
	}
	fmt.Printf("Read: %v from %v on %v\n", receivedMsg.GetContent(), receivedMsg.GetSender(), conn.RemoteAddr())
	return receivedMsg, nil
}

func (cs *ChatServer) SendResponse(conn net.Conn, msg gochat.ChatMessage) {
	sender := "server"
	sendTime := uint64(time.Now().Unix())
	fmt.Printf("Sending response: %s\n", msg)
	var response = gochat.ChatMessage{
		Sender:   &sender,
		SendTime: &sendTime,
		Content:  &msg,
	}
	msgBytes, err3 := proto.Marshal(&response)
	if gochatutil.CheckError(err3) {
		return
	}
	_, err4 := conn.Write(msgBytes)
	if gochatutil.CheckError(err4) {
		return
	}
}

// handleConnection: reads requests from clients and sends responses
func (cs *ChatServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		receivedMsg, err := cs.ReadMessage(conn)
		if err != nil {
			return
		}
		newMsg := strings.ToUpper(receivedMsg)
		cs.SendResponse(conn, newMsg)
	}
}

func parseAddress(addr string, port string) (*net.TCPAddr, error) {
	var addrBuf bytes.Buffer
	if addr != "" {
		addrBuf.WriteString(addr)
	}
	addrBuf.WriteString(":")
	addrBuf.WriteString(port)
	return net.ResolveTCPAddr("tcp", addrBuf.String())
}

func main() {
	host := flag.String("host", "localhost", "The host to listen on")
	port := flag.String("port", "8083", "The port to listen on")
	flag.Parse()
	cs := ChatServer{}
	cs.start(*host, *port)
}
