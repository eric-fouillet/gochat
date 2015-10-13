package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/eric-fouillet/gochat"
	"github.com/golang/protobuf/proto"
)

// ChatServer represents a chat server
type ChatServer struct {
	Address  *net.TCPAddr
	Name     string
	BindAddr net.TCPAddr
}

// Starts listening on the address and port given in parameters
func (cs *ChatServer) start(addr string, port string) (err error) {
	cs.Address, err = parseAddress(addr, port)
	if err != nil {
		fmt.Printf("Error while resolving address: <%v> from %v", err, cs.Address.String())
		return err
	}
	listener, err2 := net.ListenTCP("tcp", cs.Address)
	if err2 != nil {
		fmt.Printf("Error while starting server: %v\n", err2)
		return err2
	}
	fmt.Printf("Listening on address %v\n", cs.Address.String())
	for {
		conn, _ := listener.Accept()
		fmt.Printf("Accepted connection from %v !\n", conn.RemoteAddr().String())
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		buf := make([]byte, 1024)
		readBytes, err := bufio.NewReader(conn).Read(buf)
		//sz, err := binary.ReadVarint(input)
		if err != nil {
			fmt.Printf("lost connection with %v\n", conn.RemoteAddr())
			return
		}
		var receivedMsg = new(gochat.ChatMessage)
		err2 := proto.Unmarshal(buf[:readBytes], receivedMsg)
		if err2 != nil {
			log.Fatal("unmarshaling error: ", err2)
		}
		fmt.Printf("Read: %v from %v on %v\n", receivedMsg.GetContent(), receivedMsg.GetSender(), conn.RemoteAddr())
		newMsg := strings.ToUpper(receivedMsg.GetContent())
		conn.Write([]byte(newMsg + "\n"))
	}
}

func parseAddress(addr string, port string) (*net.TCPAddr, error) {
	var addrBuf bytes.Buffer
	if addr != "" {
		addrBuf.WriteString(addr)
	}
	addrBuf.WriteString(":")
	/*portBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(portBuf, port)*/
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
