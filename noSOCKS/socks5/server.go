package socks5

import (
	"fmt"
	"net"
)

func StartServer(portNum int) error {

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", portNum))
	if err != nil {
		return fmt.Errorf("Error starting server: %v\n", err)
	}

	fmt.Println("Proxy listening on port ", portNum)

	for {
		conn, err := listen.Accept()
		if err != nil {
			return fmt.Errorf("Accept error: %v\n", err)
		}
		go HandleClient(conn, portNum)
	}
}
