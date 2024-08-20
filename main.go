package main

import (
	"flag"
	"fmt"
	"os"
	"slices"

	tunnel "github.com/tarikcarli/go-tunnel/modes"
)

func main() {
	args := tunnel.NewTunnelArgs()
	flag.StringVar(&(args.Mode), "mode", "tunnel", "mode can be one of [tunnel, client, server]")
	flag.StringVar(&(args.Host), "host", "", "only used server mode, listen host address to serve tunnel request.")
	flag.IntVar(&(args.MinIdleConnection), "mid-idle-connection", 10, "only used reverse tunnel setup, client preopen connection to server to reduce initial latency.")
	flag.StringVar(&(args.Secret), "secret", "", "it is only used  when encrypt-tunnel is true. Make tunnel communication encrypted with symmetric encription by using pre shared secret.")
	flag.StringVar(&(args.Server), "server", "", "Server address, example: localhost:5602")
	flag.StringVar(&(args.Target), "target", "", "Target address, example: localhost:5602")
	flag.StringVar(&(args.Source), "source", "", "Source address, example: localhost:5602")
	fmt.Printf("args: %+v\n", args)
	validate(args)
	if args.Mode == "tunnel" {
		tunnel.MakeTunnel(args)
	} else if args.Mode == "client" {
		tunnel.MakeClient(args)
	} else if args.Mode == "server" {
		tunnel.MakeServer(args)
	}
}

func validate(args tunnel.TunnelArgs) {
	if !slices.Contains([]string{"tunnel", "server", "client"}, args.Mode) {
		fmt.Println("mode property is required and its values must be one of tunnel, client or server")
		os.Exit(1)
	}

	if args.Mode == "Server" && args.Host == "" {
		fmt.Println("host property is required when mode property is server, and its values should comply to <bind-address>:<port> format.")
		os.Exit(1)
	}

	if args.Secret == "" {
		fmt.Println("secret property must be always set. It is used by control tcp connection to authenticate clients.")
		os.Exit(1)
	}

	if args.Mode == "Client" && args.Server == "" {
		fmt.Println("server property is required when mode property is server, and its values should comply to <bind-address>:<port> format.")
		os.Exit(1)
	}

	if args.Mode == "Client" && args.Source == "" {
		fmt.Println("Source property is required when mode property is server, and its values should comply to <bind-address>:<port> format.")
		os.Exit(1)
	}

	if args.Mode == "Client" && args.Target == "" {
		fmt.Println("Target property is required when mode property is server, and its values should comply to <bind-address>:<port> format.")
		os.Exit(1)
	}
}
