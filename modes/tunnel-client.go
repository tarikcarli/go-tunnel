package tunnel

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

func MakeClient() {
	conn, err := net.Dial("tcp", Args.Server)
	if err != nil {
		fmt.Printf("Couldn't connect server machine error:%+v\n", err)
		os.Exit(1)
	}
	message := make(map[string]string)
	message["type"] = "CONFIGURE_CLIENT"
	message["target"] = Args.Target
	sendBuf, err := json.Marshal(message)
	contentLength := [4]byte{}
	binary.LittleEndian.PutUint32(contentLength[:], uint32(len(sendBuf)))
	conn.Write(contentLength[:])
	if err != nil {
		fmt.Printf("conn.Write error: %+v\n", err)
		os.Exit(1)
	}
	conn.Write(sendBuf)
	if err != nil {
		fmt.Printf("conn.Write error: %+v\n", err)
		os.Exit(1)
	}
	for {
		buf := [2048]byte{}
		readSize, err := conn.Read(buf[0:4])
		if err != nil {
			fmt.Printf("conn.Read error: %+v\n", err)
			os.Exit(1)
		}
		contentLength := binary.LittleEndian.Uint32(buf[0:4])
		readSize = 0
		for readSize < int(contentLength) {
			chunkReadSize, err := conn.Read(buf[readSize:contentLength])
			if err != nil {
				fmt.Printf("conn.Read error: %+v\n", err)
				os.Exit(1)
			}
			readSize += chunkReadSize
		}
		message := make(map[string]string)
		json.Unmarshal(buf[:contentLength], &message)
		if err != nil {
			fmt.Printf("json.Unmarshal error: %+v\n", err)
			os.Exit(1)
		}
		fmt.Printf("map %+v\n", message)
		if message["type"] == "MAKE_NEW_CONNECTION" {
			incomingConn, err := net.Dial("tcp", Args.Target)
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
	outgoingConn, err := net.Dial("tcp", Args.Source)
	if err != nil {
		fmt.Printf("createTunnelConn net.Dial Source error: %+v\n", err)
		panic(err)
	}
	outgoingConn.Write(buffer[:readSize])
	go handleTunnelConn(incomingConn, outgoingConn, nil)
	go handleTunnelConn(outgoingConn, incomingConn, nil)
}
