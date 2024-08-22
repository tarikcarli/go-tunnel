package tunnel

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/tarikcarli/go-tunnel/utils"
)

type ClientConfig struct {
	listerner      net.Listener
	controllerConn net.Conn
	idleConns      []net.Conn
}

var clients []ClientConfig

func MakeServer() {
	listener, err := net.Listen("tcp", utils.Args.Host)
	if err != nil {
		fmt.Printf("net.Listen error: %+v\n", err)
		os.Exit(1)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("listener.Accept error: %+v\n", err)
		}
		fmt.Printf("%s Accept, connection: %+v\n", utils.Args.Host, conn.RemoteAddr())
		go handleControllerConn(conn)
	}
}

func handleControllerConn(conn net.Conn) {
	var listener net.Listener = nil
	var err error = nil
	var errChan = make(chan error)

	defer func() {
		err := recover()
		fmt.Printf("handleControllerConn defer error: %+v\n", err)
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
	conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	message := make(map[string]string)
	utils.ReadConn(conn, &message)
	if message["type"] != "CONFIGURE_CLIENT" {
		fmt.Printf("PROTOCOL ERROR message %+v\n", message)
		panic(err)
	}
	listener, err = net.Listen("tcp", message["target"])
	if err != nil {
		fmt.Printf("net.Listen error: %+v\n", err)
		panic(err)
	}
	client := ClientConfig{listerner: listener, controllerConn: conn, idleConns: make([]net.Conn, 0, 100)}
	clients = append(clients, client)
	for i := 0; utils.Args.MinIdleConnection-len(client.idleConns) > i; i++ {
		sendNewConn(conn, client)
	}

	go func() {
		defer func() {
			err := recover()
			switch v := err.(type) {
			case error:
				errChan <- v
			default:
				errChan <- errors.New("unknown recover")
			}
		}()
		for {
			conn.SetReadDeadline(time.Now().Add(time.Hour * 24 * 365 * 10))
			message := make(map[string]string)
			utils.ReadConn(conn, &message)
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
			fmt.Printf("%s Accept, connection: %+v\n", message["target"], conn.RemoteAddr())
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
					go utils.HandleTunnelConn(tunnelConn, idleConn, nil)
					go utils.HandleTunnelConn(idleConn, tunnelConn, nil)
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
	utils.WriteConn(conn, message)
}
