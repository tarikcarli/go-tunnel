package tunnel

import (
	"fmt"
	"net"
)

func MakeClient(args TunnelArgs) {
	conn, err := net.Dial("tcp", args.Server)
	if err != nil {
		fmt.Printf("Couldn't connect server machine error:%+v\n", err)
	}
}
