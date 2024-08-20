package tunnel

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type ClientConfig struct {
	listerner      net.Listener
	controllerConn net.Conn
	idleConns      []net.Conn
}

var clients []ClientConfig

func MakeServer() {
	listener, err := net.Listen("tcp", Args.Host)
	if err != nil {
		fmt.Printf("net.Listen error: %+v\n", err)
		os.Exit(1)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("listener.Accept error: %+v\n", err)
		}
		fmt.Printf("%s Accept, connection: %+v\n", Args.Host, conn.RemoteAddr())
		go handleControllerConn(conn, Args)
	}
}

func handleControllerConn(conn net.Conn, args TunnelArgs) {
	var listener net.Listener = nil
	var err error = nil
	var errChan = make(chan error)

	defer func() {
		err := recover()
		close(errChan)
		errChan = nil
		if listener != nil {
			err = listener.Close()
			if err != nil {
				fmt.Printf("listener.Close error: %+v\n", err)
			}
		}
		err = conn.Close()
		if err != nil {
			fmt.Printf("conn.Close error: %+v\n", err)
		}
		var index int = -1
		for i, client := range clients {
			if strings.Split(client.controllerConn.RemoteAddr().String(), ":")[0] == strings.Split(conn.RemoteAddr().String(), ":")[0] {
				index = i
				break
			}
		}
		if index != -1 {
			clients = append(clients[:index], clients[index+1:]...)
		}
	}()
	var buffer [2048]byte
	conn.SetReadDeadline(time.Now().Add(time.Second * 5))

	readSize, err := conn.Read(buffer[:4])
	if err != nil {
		fmt.Printf("conn.Read error: %+v\n", err)
		panic(err)
	}
	readSize = 0
	contentLength := binary.LittleEndian.Uint32(buffer[0:4])
	for readSize < int(contentLength) {
		chunkReadSize, err := conn.Read(buffer[readSize:contentLength])
		if err != nil {
			fmt.Printf("conn.Read error: %+v\n", err)
			panic(err)
		}
		readSize += chunkReadSize
	}

	payload := make(map[string]string)
	err = json.Unmarshal(buffer[:contentLength], &payload)
	if err != nil {
		fmt.Printf("json.Unmarshal error: %+v\n", err)
		panic(err)
	}
	if payload["type"] != "CONFIGURE_CLIENT" {
		fmt.Printf("PROTOCOL ERROR message %+v\n", payload)
		panic(err)
	}
	fmt.Printf("payload: %+v\n", payload)
	listener, err = net.Listen("tcp", payload["target"])
	if err != nil {
		fmt.Printf("net.Listen error: %+v\n", err)
		panic(err)
	}
	client := ClientConfig{listerner: listener, controllerConn: conn, idleConns: make([]net.Conn, 0, 100)}
	clients = append(clients, client)
	for i := 0; args.MinIdleConnection-len(client.idleConns) > i; i++ {
		sendNewConn(conn, client)
	}

	go func() {
		buffer := [1024]byte{}
		for {
			conn.SetReadDeadline(time.Now().Add(time.Hour * 24 * 365 * 10))
			_, err = conn.Read(buffer[:])
			if err != nil {
				fmt.Printf("conn.Read error: %+v\n", err)
				if errChan != nil {
					errChan <- err
					return
				}
			}
		}
	}()

	go func() {
		for {
			tunnelConn, err := listener.Accept()
			if err != nil {
				fmt.Printf("listener.Accept error: %+v\n", err)
				if errChan != nil {
					errChan <- err
				}
				return
			}
			fmt.Printf("%s Accept, connection: %+v\n", payload["target"], conn.RemoteAddr())
			if strings.Split(tunnelConn.RemoteAddr().String(), ":")[0] == strings.Split(conn.RemoteAddr().String(), ":")[0] {
				client.idleConns = append(client.idleConns, tunnelConn)
			} else {
				if len(client.idleConns) == 0 {
					err := tunnelConn.Close()
					if err != nil {
						fmt.Printf("tunnelConn.Close error: %+v\n", err)
					}
				} else {
					idleConn := client.idleConns[0]
					client.idleConns = append(client.idleConns[:0], client.idleConns[1:]...)
					sendNewConn(conn, client)
					go handleTunnelConn(tunnelConn, idleConn, nil)
					go handleTunnelConn(idleConn, tunnelConn, nil)
				}
			}
		}
	}()
	err = <-errChan
	panic(err)
}

func sendNewConn(conn net.Conn, client ClientConfig) {
	message := make(map[string]string)
	message["type"] = "MAKE_NEW_CONNECTION"
	buf, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("json.Marshal error:%+v\n", err)
		panic(err)
	}
	contentLength := [4]byte{}
	binary.LittleEndian.PutUint32(contentLength[:], uint32(len(buf)))
	_, err = conn.Write(contentLength[:])
	if err != nil {
		panic(err)
	}
	_, err = conn.Write(buf)
	if err != nil {
		panic(err)
	}
}
