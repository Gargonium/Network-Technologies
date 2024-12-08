package main

import (
	"fmt"
	"noSOCKS/socks5"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <port>")
		return
	}

	port := os.Args[1]
	portNum, err := strconv.Atoi(port)
	if err != nil {
		fmt.Println("Invalid port number:", err)
		return
	}

	err = socks5.StartServer(portNum)
	if err != nil {
		fmt.Println("Server cannot start:", err)
		return
	}
}
