package tunnel

import (
	"fmt"
	"net"
	"os"

	"github.com/tarikcarli/go-tunnel/utils"
)

func MakeClient() {
	conn, err := net.Dial("tcp", utils.Args.Server)
	if err != nil {
		fmt.Printf("Couldn't connect server machine error:%+v\n", err)
		os.Exit(1)
	}
	message := make(map[string]string)
	message["type"] = "CONFIGURE_CLIENT"
	message["target"] = utils.Args.Target
	utils.WriteConn(conn, message)
	if err != nil {
		fmt.Printf("conn.Write error: %+v\n", err)
		os.Exit(1)
	}
	for {
		message := make(map[string]string)
		utils.ReadConn(conn, &message)
		if message["type"] == "MAKE_NEW_CONNECTION" {
			incomingConn, err := net.Dial("tcp", utils.Args.Target)
			fmt.Printf("net.Dial %s\n", incomingConn.RemoteAddr().String())
			if err != nil {
				fmt.Printf("createTunnelConn net.Dial Target error: %+v\n", err)
			} else {
				go handleIncomingConn(incomingConn)
			}
		}
	}
}

func handleIncomingConn(incomingConn net.Conn) {
	defer func() {
		err := recover()
		if err != nil {
			err = incomingConn.Close()
			if err != nil {
				fmt.Printf("incomingConn.Close error: %+v\n")
			}
		}
	}()
	buffer := [1024]byte{}
	readSize, err := incomingConn.Read(buffer[:])
	if err != nil {
		fmt.Printf("incomingConn.Read error: %+v\n", err)
		panic(err)
	}
	outgoingConn, err := net.Dial("tcp", utils.Args.Source)
	if err != nil {
		fmt.Printf("createTunnelConn net.Dial Source error: %+v\n", err)
		panic(err)
	}
	outgoingConn.Write(buffer[:readSize])
	go utils.HandleTunnelConn(incomingConn, outgoingConn, nil)
	go utils.HandleTunnelConn(outgoingConn, incomingConn, nil)
}
