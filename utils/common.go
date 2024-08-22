package utils

import (
	"crypto/cipher"
	"encoding/binary"
	"encoding/json"
	"errors"
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
var Block cipher.Block = nil

func HandleTunnelConn(readConn net.Conn, writeConn net.Conn, finish chan int) {
	defer func() {
		err := recover()
		if err != nil {
			switch err := err.(type) {
			case error:
				err = readConn.Close()
				if err != nil && !errors.Is(err, net.ErrClosed) {
					fmt.Printf("readConn.Close error: %+v\n", err)
				}
				err = writeConn.Close()
				if err != nil && !errors.Is(err, net.ErrClosed) {
					fmt.Printf("writeConn.Close error: %+v\n", err)
				}
				if finish != nil {
					finish <- 1
				}
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

func WriteConn(conn net.Conn, message map[string]string) int {
	var messageSize int
	var buffer []byte = nil
	payload, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("json.Marshal error: %+v", err)
	}
	contentLength := len(payload)

	if Block != nil {
		paddingByte := (contentLength + 4) % Block.BlockSize()
		contentLength := contentLength
		buffer = make([]byte, 0, contentLength+4+paddingByte)
		binary.LittleEndian.PutUint32(buffer[0:4], uint32(contentLength))
		buffer = buffer[:4]
		buffer = append(buffer, payload...)
		buffer = append(buffer, make([]byte, paddingByte)...)
		Block.Encrypt(buffer, buffer)
	} else {
		contentLength := contentLength
		buffer = make([]byte, 0, contentLength+4)
		binary.LittleEndian.PutUint32(buffer[0:4], uint32(contentLength))
		buffer = buffer[:4]
		buffer = append(buffer, payload...)
	}

	messageSize = len(buffer)
	writeSize := 0
	for writeSize < messageSize {
		n, err := conn.Write(buffer[writeSize:messageSize])
		if err != nil {
			fmt.Printf("conn error:%+v\n", err)
			panic(err)
		}
		writeSize += n
	}

	return messageSize
}

func ReadConn(conn net.Conn, message *map[string]string) {
	var contentLength int
	var messageLength int
	var buffer []byte = nil

	if Block != nil {
		firstBlock := make([]byte, Block.BlockSize(), Block.BlockSize())
		readSize := 0
		for readSize < Block.BlockSize() {
			n, err := conn.Read(firstBlock)
			if err != nil {
				fmt.Printf("conn.Read error: %+v\n", err)
				panic(err)
			}
			readSize += n
		}
		firstBlockDecrpted := make([]byte, Block.BlockSize(), Block.BlockSize())
		Block.Decrypt(firstBlockDecrpted, firstBlock)
		contentLength = int(binary.LittleEndian.Uint32(firstBlockDecrpted[0:4]))
		messageLength = contentLength + 4 + (contentLength+4)%Block.BlockSize()
		buffer = make([]byte, 0, messageLength)
		_ = append(buffer, firstBlock...)
		for readSize < messageLength {
			n, err := conn.Read(buffer[readSize:messageLength])
			if err != nil {
				fmt.Printf("conn.Read error: %+v\n", err)
				panic(err)
			}
			readSize += n
		}
		Block.Decrypt(buffer[:messageLength], buffer[:messageLength])
	} else {
		size := make([]byte, 4, 4)
		readSize := 0
		for readSize < 4 {
			n, err := conn.Read(size)
			if err != nil {
				fmt.Printf("conn.Read error: %+v\n", err)
				panic(err)
			}
			readSize += n
		}
		contentLength = int(binary.LittleEndian.Uint32(size))
		messageLength = contentLength + 4
		buffer = make([]byte, contentLength+4, contentLength+4)
		readSize = 4
		for readSize < contentLength {
			n, err := conn.Read(buffer[readSize : contentLength+4])
			if err != nil {
				fmt.Printf("conn.Read error: %+v\n", err)
				panic(err)
			}
			readSize += n
		}
	}

	json.Unmarshal(buffer[4:contentLength+4], &message)
}
