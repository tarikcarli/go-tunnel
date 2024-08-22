package tunnel

import (
	"fmt"
	"net"

	"github.com/tarikcarli/go-tunnel/utils"
)

func MakeTunnel() {
	listener, err := net.Listen("tcp", utils.Args.Host)
	if err != nil {
		fmt.Printf("net.Listen error: %+v\n", err)
	}
	for {
		incomingConn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error occured while accepting connection, error: %+v\n", err)
			continue
		}
		fmt.Printf("Accept, connection: %+v\n", incomingConn.RemoteAddr())
		outgoingConn, err := net.Dial("tcp", utils.Args.Source)
		if err != nil {
			fmt.Printf("net.Dial error: %+v\n", err)
			err = incomingConn.Close()
			if err != nil {
				fmt.Print("incomingConn.Close error: %+v\n", err)
			}
			continue
		}
		go utils.HandleTunnelConn(incomingConn, outgoingConn, nil)
		go utils.HandleTunnelConn(outgoingConn, incomingConn, nil)
	}
}
