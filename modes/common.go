package tunnel

import (
	"bytes"
	"encoding/gob"
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

type SupervisorMessageType int

const (
	createConn SupervisorMessageType = iota
)

func (w SupervisorMessageType) String() string {
	return [...]string{"Create Connection"}[w]
}
func (w SupervisorMessageType) EnumIndex() int {
	return int(w)
}

type SupervisorCommand int

const (
	AddTunnel SupervisorCommand = iota
	RemoveTunnel
	WriteIncomingConn
	WriteOutgoingConn
)

func (w SupervisorCommand) String() string {
	return [...]string{"Remove", "Close", "Write"}[w]
}
func (w SupervisorCommand) EnumIndex() int {
	return int(w)
}

type SupervisorMessage struct {
	command SupervisorCommand
	tunnel  Tunnel
}

type Tunnel struct {
	incoming net.Conn
	outgoing net.Conn
}

type ClientConfigureMessage struct {
	Target string
}

func (m *ClientConfigureMessage) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	encoder.Encode(m.Target)

	return w.Bytes(), nil
}

func (m *ClientConfigureMessage) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	decoder.Decode(&m.Target)
	return nil
}

func main() {
	d := ClientConfigureMessage{Target: "1.2.3.4:8080"}
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	encoder.Encode(d)
	buffer = bytes.NewBuffer(buffer.Bytes())
	e := new(ClientConfigureMessage)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(e)
	fmt.Println(e)
}
