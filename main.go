package main

import (
	"crypto/aes"
	"flag"
	"fmt"
	"os"
	"slices"

	tunnel "github.com/tarikcarli/go-tunnel/modes"
	"github.com/tarikcarli/go-tunnel/utils"
)

func main() {
	flag.StringVar(&(utils.Args.Mode), "mode", "tunnel", "mode can be one of [tunnel, client, server]")
	flag.StringVar(&(utils.Args.Host), "host", "", "only used server mode, listen host address to serve tunnel request.")
	flag.IntVar(&(utils.Args.MinIdleConnection), "min-idle-connections", 10, "only used reverse tunnel setup, client preopen connection to server to reduce initial latency.")
	flag.StringVar(&(utils.Args.Secret), "secret", "", "it is only used  when encrypt-tunnel is true. Make tunnel communication encrypted with symmetric encription by using pre shared secret.")
	flag.StringVar(&(utils.Args.Server), "server", "", "Server address, example: localhost:5602")
	flag.StringVar(&(utils.Args.Target), "target", "", "Target address, example: localhost:5602")
	flag.StringVar(&(utils.Args.Source), "source", "", "Source address, example: localhost:5602")
	flag.Parse()
	fmt.Printf("\n\nutils.Args: %+v\n\n\n", utils.Args)
	validate()
	if utils.Args.Secret != "" {
		block, err := aes.NewCipher([]byte(utils.Args.Secret)[0:32])
		if err != nil {
			fmt.Printf("aes.NewCipher error:%+v\n", err)
			os.Exit(1)
		}
		utils.Block = block
	}
	if utils.Args.Mode == "tunnel" {
		tunnel.MakeTunnel()
	} else if utils.Args.Mode == "client" {
		tunnel.MakeClient()
	} else if utils.Args.Mode == "server" {
		tunnel.MakeServer()
	}
}

func validate() {
	if !slices.Contains([]string{"tunnel", "server", "client"}, utils.Args.Mode) {
		fmt.Println("mode property is required and its values must be one of tunnel, client or server")
		os.Exit(1)
	}

	if utils.Args.Mode == "Server" && utils.Args.Host == "" {
		fmt.Println("host property is required when mode property is server, and its values should comply to <bind-address>:<port> format.")
		os.Exit(1)
	}

	if utils.Args.Mode == "Client" && utils.Args.Server == "" {
		fmt.Println("server property is required when mode property is server, and its values should comply to <bind-address>:<port> format.")
		os.Exit(1)
	}

	if utils.Args.Mode == "Client" && utils.Args.Source == "" {
		fmt.Println("Source property is required when mode property is server, and its values should comply to <bind-address>:<port> format.")
		os.Exit(1)
	}

	if utils.Args.Mode == "Client" && utils.Args.Target == "" {
		fmt.Println("Target property is required when mode property is server, and its values should comply to <bind-address>:<port> format.")
		os.Exit(1)
	}
}
