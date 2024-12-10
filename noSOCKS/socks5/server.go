package socks5

import (
	"fmt"
	"github.com/lesismal/nbio"
	"log"
)

// Client Greeting
// Server Choice
// Client Connection request
// Response packet from server

// SOCKS5 address

func StartServer(portNum int) error {

	engine := nbio.NewEngine(nbio.Config{
		Network: "tcp",
		Addrs:   []string{fmt.Sprintf(":%d", portNum)},
	})

	engine.OnOpen(func(c *nbio.Conn) {
		log.Printf("Новый клиент подключился: %s\n", c.RemoteAddr().String())
		HandleClient(c, portNum)
	})

	engine.OnClose(func(c *nbio.Conn, err error) {
		c.Close()
		log.Printf("Close Connection: %s\n", c.RemoteAddr().String())
	})

	err := engine.Start()

	if err != nil {
		return fmt.Errorf("Ошибка при запуске сервера: %v\n", err)
	}
	defer engine.Stop()

	<-make(chan int)

	return nil
}
