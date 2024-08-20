package tunnel

import (
	"fmt"
	"net"
)

func MakeTunnel(args TunnelArgs) {
	tunnels := make([]Tunnel, 100)
	supervisorChanel := make(chan SupervisorMessage, 10)
	go listenServer(supervisorChanel, args)
	for {
		message := <-supervisorChanel

		if message.command == AddTunnel {
			tunnels = append(tunnels, message.tunnel)
		}
		if message.command == RemoveTunnel {
			for i, tunnel := range tunnels {
				if tunnel == message.tunnel {
					if err := tunnel.incoming.Close(); err != nil {
						fmt.Printf("tunnel.incoming.Close error:%+v\n", err)
					}
					if err := tunnel.outgoing.Close(); err != nil {
						fmt.Printf("tunnel.outgoing.Close error:%+v\n", err)
					}
					tunnels = append(tunnels[:i], tunnels[i+1:]...)
				}
			}
		}
	}
}

func listenServer(supervisorChanel chan SupervisorMessage, args TunnelArgs) {
	listener, err := net.Listen("tcp", args.Host)
	if err != nil {
		fmt.Printf("net.Listen error: %+v\n", err)
	}
	for {
		incomingConn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error occured while accepting connection, error: %+v\n", err)
		}
		fmt.Printf("Accept, connection: %+v\n", incomingConn.RemoteAddr())
		outgoingConn, err := net.Dial("tcp", args.Source)
		if err != nil {
			fmt.Printf("Error occured to connect source adress, error: %+v\n", err)
			return
		}
		currentTunnel := Tunnel{incoming: incomingConn, outgoing: outgoingConn}
		supervisorChanel <- SupervisorMessage{command: AddTunnel, tunnel: currentTunnel}
		go handleTunnelConn(supervisorChanel, currentTunnel, currentTunnel.incoming, currentTunnel.outgoing)
		go handleTunnelConn(supervisorChanel, currentTunnel, currentTunnel.outgoing, currentTunnel.incoming)
	}
}

func handleTunnelConn(supervisorChanel chan SupervisorMessage, tunnel Tunnel, readConn net.Conn, writeConn net.Conn) {
	var buffer [2048]byte
	for {
		numberOfReadBytes, err := readConn.Read(buffer[:])
		if err != nil {
			fmt.Printf("readConn.Read error:%+v\n", err)
		}
		if err != nil {
			supervisorChanel <- SupervisorMessage{command: RemoveTunnel, tunnel: tunnel}
			return
		}
		numberOfWriteBytes := 0
		for numberOfWriteBytes < numberOfReadBytes {
			n, err := writeConn.Write(buffer[numberOfWriteBytes:numberOfReadBytes])
			if err != nil {
				fmt.Printf("writeConn.Write error:%+v\n", err)
				supervisorChanel <- SupervisorMessage{command: RemoveTunnel, tunnel: tunnel}
				return
			}
			numberOfWriteBytes += n
		}
	}
}
