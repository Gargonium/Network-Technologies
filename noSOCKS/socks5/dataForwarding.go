package socks5

import (
	"io"
	"log"
	"net"
)

func EstablishConnection(client net.Conn, server net.Conn) {
	defer func(server net.Conn) {
		err := server.Close()
		if err != nil {
			log.Println("Error closing server", err)
		}
	}(server)
	defer func(client net.Conn) {
		err := client.Close()
		if err != nil {
			log.Println("Error closing client", err)
		}
	}(client)

	go func() {
		for {
			if _, err := io.Copy(server, client); err != nil {
				log.Printf("Ошибка копирования из server в client: %v", err)
			}
		}
	}()

	go func() {
		for {
			if _, err := io.Copy(client, server); err != nil {
				log.Printf("Ошибка копирования из client в server: %v", err)
			}
		}
	}()

	select {}
}
