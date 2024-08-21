package main

import (
	"crypto/aes"
	"flag"
	"fmt"
	"os"
	"slices"

	tunnel "github.com/tarikcarli/go-tunnel/modes"
)

func main() {
	flag.StringVar(&(tunnel.Args.Mode), "mode", "tunnel", "mode can be one of [tunnel, client, server]")
	flag.StringVar(&(tunnel.Args.Host), "host", "", "only used server mode, listen host address to serve tunnel request.")
	flag.IntVar(&(tunnel.Args.MinIdleConnection), "min-idle-connections", 10, "only used reverse tunnel setup, client preopen connection to server to reduce initial latency.")
	flag.StringVar(&(tunnel.Args.Secret), "secret", "", "it is only used  when encrypt-tunnel is true. Make tunnel communication encrypted with symmetric encription by using pre shared secret.")
	flag.StringVar(&(tunnel.Args.Server), "server", "", "Server address, example: localhost:5602")
	flag.StringVar(&(tunnel.Args.Target), "target", "", "Target address, example: localhost:5602")
	flag.StringVar(&(tunnel.Args.Source), "source", "", "Source address, example: localhost:5602")
	flag.Parse()
	fmt.Printf("\n\ntunnel.Args: %+v\n\n\n", tunnel.Args)
	validate()
	if tunnel.Args.Secret != "" {
		block, err := aes.NewCipher([]byte(tunnel.Args.Secret)[0:32])
		if err != nil {
			fmt.Printf("aes.NewCipher error:%+v\n", err)
			os.Exit(1)
		}
		tunnel.Block = block
	}
	if tunnel.Args.Mode == "tunnel" {
		tunnel.MakeTunnel()
	} else if tunnel.Args.Mode == "client" {
		tunnel.MakeClient()
	} else if tunnel.Args.Mode == "server" {
		tunnel.MakeServer()
	}
}

func validate() {
	if !slices.Contains([]string{"tunnel", "server", "client"}, tunnel.Args.Mode) {
		fmt.Println("mode property is required and its values must be one of tunnel, client or server")
		os.Exit(1)
	}

	if tunnel.Args.Mode == "Server" && tunnel.Args.Host == "" {
		fmt.Println("host property is required when mode property is server, and its values should comply to <bind-address>:<port> format.")
		os.Exit(1)
	}

	if tunnel.Args.Mode == "Client" && tunnel.Args.Server == "" {
		fmt.Println("server property is required when mode property is server, and its values should comply to <bind-address>:<port> format.")
		os.Exit(1)
	}

	if tunnel.Args.Mode == "Client" && tunnel.Args.Source == "" {
		fmt.Println("Source property is required when mode property is server, and its values should comply to <bind-address>:<port> format.")
		os.Exit(1)
	}

	if tunnel.Args.Mode == "Client" && tunnel.Args.Target == "" {
		fmt.Println("Target property is required when mode property is server, and its values should comply to <bind-address>:<port> format.")
		os.Exit(1)
	}
}
