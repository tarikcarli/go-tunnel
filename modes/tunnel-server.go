package tunnel

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type ClientConfig struct {
	listerner      net.Listener
	controllerConn net.Conn
	idleConns      []net.Conn
	incomingConns  []net.Conn
	outgoingConns  []net.Conn
}

var clients []ClientConfig

func MakeServer(args TunnelArgs) {
	listener, err := net.Listen("tcp", args.Host)
	if err != nil {
		fmt.Printf("net.Listen error: %+v\n", err)
		os.Exit(1)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("listener.Accept error: %+v\n", err)
		}
		go handleControllerConn(conn, args)
	}
}

func handleControllerConn(conn net.Conn, args TunnelArgs) error {
	var buffer [2048]byte
	conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	n, err := conn.Read(buffer[:])
	if err != nil {
		fmt.Printf("conn.Read error: %+v\n", err)
		return conn.Close()
	}
	size, err := strconv.ParseInt(string(buffer[:4]), 10, 32)
	if err != nil {
		fmt.Printf("strconv.ParseInt error: %+v\n", err)
		conn.Close()
	}
	totalRead := n
	for size > int64(totalRead) {
		n, err := conn.Read(buffer[totalRead:])
		if err != nil {
			fmt.Printf("conn.Read error: %+v\n", err)
			return conn.Close()
		}
		totalRead += n
	}
	decoder := gob.NewDecoder(bytes.NewBuffer(buffer[4:totalRead]))
	clientConf := ClientConfigureMessage{}
	err = decoder.Decode(&clientConf)
	if err != nil {
		fmt.Printf("decoder.Decode error: %+v\n", err)
		return conn.Close()
	}
	found := false
	for _, client := range clients {
		if strings.Split(client.controllerConn.RemoteAddr().String(), ":")[0] == strings.Split(conn.RemoteAddr().String(), ":")[0] {
			err := client.controllerConn.Close()
			if err != nil {
				fmt.Printf("client.controllerConn.Close error: %+v\n", err)
			}
			client.controllerConn = conn
			found = true
		}
	}
	listener, err := net.Listen("tcp", clientConf.Target)
	if err != nil {
		fmt.Printf("net.Listen error: %+v\n", err)
		return conn.Close()
	}
	var client *ClientConfig = nil
	if !found {
		client = &ClientConfig{listerner: listener, controllerConn: conn, idleConns: make([]net.Conn, 10), incomingConns: make([]net.Conn, 10), outgoingConns: make([]net.Conn, 10)}
		clients = append(clients, *client)
	}
	for {
		for i := 0; args.MinIdleConnection > i; i++ {
			// conn.Write() // populate idleConns
		}
		tunnelConn, err := listener.Accept()
		if err != nil {
			fmt.Printf("listener.Accept error: %+v\n", err)
		}
		if strings.Split(tunnelConn.RemoteAddr().String(), ":")[0] == strings.Split(conn.RemoteAddr().String(), ":")[0] {
			client.idleConns = append(client.idleConns, tunnelConn)
		} else {
			if len(client.idleConns) == 0 {
				tunnelConn.Close()
			}
		}
	}
}

func sendMessage(conn net.Conn, messageType SupervisorMessageType) {

}
func handleClient(conn net.Conn, client *ClientConfig, args TunnelArgs) {

}
