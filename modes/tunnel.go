package tunnel

import (
	"fmt"
	"net"
)

func MakeTunnel() {
	listener, err := net.Listen("tcp", Args.Host)
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
		outgoingConn, err := net.Dial("tcp", Args.Source)
		if err != nil {
			fmt.Printf("net.Dial error: %+v\n", err)
			err = incomingConn.Close()
			if err != nil {
				fmt.Print("incomingConn.Close error: %+v\n", err)
			}
			continue
		}
		go handleTunnelConn(incomingConn, outgoingConn, nil)
		go handleTunnelConn(outgoingConn, incomingConn, nil)
	}
}
