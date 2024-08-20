package tunnel

import (
	"fmt"
	"net"
)

type TunnelArgs struct {
	Mode              string
	Host              string
	MinIdleConnection int
	Secret            string
	Server            string
	Target            string
	Source            string
}

func NewTunnelArgs() TunnelArgs {
	return TunnelArgs{}
}

var Args = NewTunnelArgs()

func handleTunnelConn(readConn net.Conn, writeConn net.Conn, finish chan int) {
	defer func() {
		err := recover()
		if err != nil {
			err = readConn.Close()
			if err != nil {
				fmt.Printf("readConn.Close error: %+v\n", err)
			}
			err = writeConn.Close()
			if err != nil {
				fmt.Printf("writeConn.Close error: %+v\n", err)
			}
			if finish != nil {
				finish <- 1
			}
		}
	}()
	var buffer [2048]byte
	for {
		readSize, err := readConn.Read(buffer[:])
		if err != nil {
			fmt.Printf("readConn.Read error:%+v\n", err)
			panic(err)
		}
		writeSize := 0
		for writeSize < readSize {
			n, err := writeConn.Write(buffer[writeSize:readSize])
			if err != nil {
				fmt.Printf("writeConn.Write error:%+v\n", err)
				panic(err)
			}
			writeSize += n
		}
	}
}
